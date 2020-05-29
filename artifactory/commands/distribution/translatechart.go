package distribution

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/release-bundle-generators/artifactory/commands/generic"
	"github.com/jfrog/release-bundle-generators/artifactory/spec"
	"github.com/jfrog/release-bundle-generators/utils/cliutils"
	"github.com/jfrog/release-bundle-generators/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"net/http"
	"sort"
	"strings"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
)

type TranslateChartCommand struct {
	rtDetails            *config.ArtifactoryDetails
	releaseBundlesParams distributionServicesUtils.ReleaseBundleParams
	sourceChartPath      string
	dockerRepo           string
	dryRun               bool
}

func NewTranslateChartCommand() *TranslateChartCommand {
	return &TranslateChartCommand{}
}

func (tc *TranslateChartCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *TranslateChartCommand {
	tc.rtDetails = rtDetails
	return tc
}

func (tc *TranslateChartCommand) SetReleaseBundleCreateParams(params distributionServicesUtils.ReleaseBundleParams) *TranslateChartCommand {
	tc.releaseBundlesParams = params
	return tc
}

func (tc *TranslateChartCommand) SetSourceChartPath(sourceChartPath string) *TranslateChartCommand {
	tc.sourceChartPath = sourceChartPath
	return tc
}

func (tc *TranslateChartCommand) SetDockerRepo(dockerRepo string) *TranslateChartCommand {
	tc.dockerRepo = dockerRepo
	return tc
}

func (tc *TranslateChartCommand) SetDryRun(dryRun bool) *TranslateChartCommand {
	tc.dryRun = dryRun
	return tc
}

func (tc *TranslateChartCommand) Run() error {
	body, err := readFileFromArtifactory(tc.rtDetails, tc.sourceChartPath)
	if err != nil {
		return err
	}
	defer body.Close()
	chrt, err := chartutil.LoadArchive(body)
	if err != nil {
		return err
	}
	specstr, expected, err := createFilespec(chrt, extractRepo(tc.sourceChartPath), tc.dockerRepo)
	if err != nil {
		return err
	}
	specfiles := new(spec.SpecFiles)
	err = json.Unmarshal([]byte(specstr), specfiles)
	if err != nil {
		return err
	}
	createBundle := CreateBundleCommand{
		rtDetails: tc.rtDetails,
		releaseBundlesParams: tc.releaseBundlesParams,
		spec: specfiles,
		dryRun: tc.dryRun,
	}
	err = createBundle.Run()
	if err != nil {
		return err
	}
	actual, err := checkExisting(tc.rtDetails, specfiles)
	if err != nil {
		return err
	}
	missing := make([]string, 0)
	fmt.Println("Found:")
	for _, name := range expected {
		line := "/" + name
		found := false
		if !strings.HasSuffix(line, ".tgz") {
			line = strings.ReplaceAll(line, ":", "/") + "/"
		}
		for _, path := range actual {
			if strings.Contains(path, line) {
				fmt.Println("- " + name)
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, name)
		}
	}
	fmt.Println("Missing:")
	if len(missing) <= 0 {
		missing = append(missing, "none")
	}
	for _, line := range missing {
		fmt.Println("- " + line)
	}
	return nil
}

func (tc *TranslateChartCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return tc.rtDetails, nil
}

func (tc *TranslateChartCommand) CommandName() string {
	return "generate_from_chart"
}

func extractRepo(path string) string {
	if path[0] == '/' {
		path = path[1:]
	}
	return strings.SplitN(path, "/", 2)[0]
}

func readFileFromArtifactory(artDetails *config.ArtifactoryDetails, downloadPath string) (io.ReadCloser, error) {
	downloadUrl := urlAppend(artDetails.Url, downloadPath)
	auth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	securityDir, err := cliutils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	client, err := rthttpclient.ArtifactoryClientBuilder().
		SetCertificatesPath(securityDir).
		SetInsecureTls(artDetails.InsecureTls).
		SetServiceDetails(&auth).
		Build()
	if err != nil {
		return nil, err
	}
	httpClientDetails := auth.CreateHttpClientDetails()
	body, resp, err := client.ReadRemoteFile(downloadUrl, &httpClientDetails)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = errorutils.CheckError(errors.New(resp.Status + " received when attempting to download " + downloadUrl))
	}
	return body, err
}

func createFilespec(chrt *chart.Chart, helmrepo, dockerrepo string) (string, []string, error) {
	flist := make([]string, 0)
	files, err := renderutil.Render(chrt, &chart.Config{Raw: "{}"}, renderutil.Options{})
	if err != nil {
		return "", flist, err
	}
	spec := "{\"files\":["
	lines := extractImages(files)
	for _, line := range sortStringMap(lines) {
		splits := strings.SplitN(line, "/", 2)
		image := splits[len(splits)-1]
		cname := strings.ReplaceAll(image, ":", "/") + "/"
		path1, _ := json.Marshal(dockerrepo+"/"+cname)
		path2, _ := json.Marshal(dockerrepo+"/*/"+cname)
		spec = spec + "{\"pattern\":" + string(path1) + "},"
		spec = spec + "{\"pattern\":" + string(path2) + "},"
		flist = append(flist, image)
	}
	deps := map[string]*chart.Chart{}
	crawlRequirements(deps, chrt)
	for _, c := range sortChartMap(deps) {
		cname := c.Metadata.Name + "-" + c.Metadata.Version + ".tgz"
		path1, _ := json.Marshal(helmrepo+"/"+cname)
		path2, _ := json.Marshal(helmrepo+"/*/"+cname)
		spec = spec + "{\"pattern\":" + string(path1) + "},"
		spec = spec + "{\"pattern\":" + string(path2) + "},"
		flist = append(flist, cname)
	}
	spec = spec[:len(spec)-1]
	spec = spec + "]}"
	return spec, flist, nil
}

func checkExisting(rtDetails *config.ArtifactoryDetails, spec *spec.SpecFiles) ([]string, error) {
	flist := make([]string, 0)
	cmd := generic.NewSearchCommand()
	cmd.SetRtDetails(rtDetails).SetSpec(spec)
	err := cmd.Search()
	if err != nil {
		return flist, err
	}
	for _, result := range cmd.SearchResult() {
		flist = append(flist, result.Path)
	}
	return flist, nil
}

func sortStringMap(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	vals := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, in[k])
	}
	return vals
}

func sortChartMap(in map[string]*chart.Chart) []*chart.Chart {
	keys := make([]string, 0, len(in))
	vals := make([]*chart.Chart, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, in[k])
	}
	return vals
}

func urlAppend(url, path string) string {
	if url[len(url)-1] != '/' {
		url = url + "/"
	}
	if path[0] == '/' {
		path = path[1:]
	}
	return url + path
}

func extractImages(files map[string]string) map[string]string {
	lines := map[string]string{}
	for _, v := range files {
		for _, line := range strings.Split(v, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "image:") {
				continue
			}
			line = strings.TrimPrefix(line, "image:")
			line = strings.TrimSpace(line)
			if len(line) <= 0 {
				continue
			}
			if line[0] == '\'' && line[len(line)-1] == '\'' ||
				line[0] == '"' && line[len(line)-1] == '"' {
				line = line[1:len(line)-1]
			}
			lines[line] = line
		}
	}
	return lines
}

func crawlRequirements(reqs map[string]*chart.Chart, chrt *chart.Chart) {
	reqs[chrt.Metadata.Name] = chrt
	for _, req := range chrt.GetDependencies() {
		crawlRequirements(reqs, req)
	}
}
