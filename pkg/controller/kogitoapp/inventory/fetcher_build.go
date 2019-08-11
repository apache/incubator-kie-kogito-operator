package inventory

import (
	"fmt"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/definitions"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	clibuildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	cliimgv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// BuildState describes the state of the build
type BuildState struct {
	ImageExists  bool
	BuildRunning bool
}

// EnsureImageBuild checks for the corresponding image for the build. If there's no image, verifies if the build still running.
// Returns a BuildState structure describing it's results
func EnsureImageBuild(buildcli clibuildv1.BuildV1Interface, imgcli cliimgv1.ImageV1Interface, bc *buildv1.BuildConfig) (*BuildState, error) {
	state := &BuildState{
		ImageExists:  false,
		BuildRunning: false,
	}
	bcNamed := types.NamespacedName{
		Name:      bc.Name,
		Namespace: bc.Namespace,
	}
	if img, err := FetchDockerImage(imgcli, bcNamed); err != nil {
		return state, err
	} else if img == nil {
		log.Infof("Image not found for build %s", bc.Name)
		state.ImageExists = false
		if running, err := BuildIsRunning(buildcli, bc); running {
			log.Infof("Build %s is still running", bc.Name)
			state.BuildRunning = true
			return state, nil
		} else if err != nil {
			return state, err
		}
		// TODO: ensure that we don't have errors in the builds and inform this to the user
		log.Debugf("There's no image and no build running or pending for %s.", bc.Name)
		return state, nil
	}
	state.ImageExists = true
	return state, nil
}

// TriggerBuild triggers a new build
func TriggerBuild(cli clibuildv1.BuildV1Interface, bc *buildv1.BuildConfig, kogitoApp *v1alpha1.KogitoApp) (bool, error) {
	if exists, err := checkBuildConfigExists(cli, bc); !exists {
		log.Warnf("Impossible to trigger a new build for %s. Not exists.", bc.Name)
		return false, err
	}
	// catch panic when FakeClient Build is unable to handle dc properly
	defer func() {
		if err := recover(); err != nil {
			log.Info("Skip build triggering duo to a bug on FakeBuild: github.com/openshift/client-go/build/clientset/versioned/typed/build/v1/fake/fake_buildconfig.go:134")
		}
	}()
	buildRequest := definitions.NewBuildRequest(kogitoApp, bc)
	build, err := cli.BuildConfigs(bc.Namespace).Instantiate(bc.Name, &buildRequest)
	if err != nil {
		return false, err
	}

	log.Info("Build triggered: ", build.Name)
	return true, nil
}

// BuildIsRunning checks if there's a build on New, Pending or Running state for the buildConfiguration
func BuildIsRunning(cli clibuildv1.BuildV1Interface, bc *buildv1.BuildConfig) (bool, error) {
	if exists, err := checkBuildConfigExists(cli, bc); !exists {
		return false, err
	}
	list, err := cli.Builds(bc.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s",
			definitions.LabelKeyAppName,
			bc.Labels[definitions.LabelKeyAppName],
			definitions.LabelKeyBuildType,
			bc.Labels[definitions.LabelKeyBuildType],
		),
		IncludeUninitialized: false,
	})
	if err != nil {
		return false, err
	}
	for _, item := range list.Items {
		// it's the build from our buildConfig
		if strings.HasPrefix(item.Name, bc.Name) {
			log.Debugf("Checking status of build '%s'", item.Name)
			if item.Status.Phase == buildv1.BuildPhaseNew ||
				item.Status.Phase == buildv1.BuildPhasePending ||
				item.Status.Phase == buildv1.BuildPhaseRunning {
				log.Debugf("Build %s is still running", item.Name)
				return true, nil
			}
			log.Debugf("Build %s status is %s", item.Name, item.Status.Phase)
		}
	}
	return false, nil
}

func checkBuildConfigExists(cli clibuildv1.BuildV1Interface, bc *buildv1.BuildConfig) (bool, error) {
	if _, err := cli.BuildConfigs(bc.Namespace).Get(bc.Name, metav1.GetOptions{}); err != nil && errors.IsNotFound(err) {
		log.Warnf("BuildConfig not found in namespace")
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
