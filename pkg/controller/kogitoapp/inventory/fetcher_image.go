package inventory

import (
	"encoding/json"
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/definitions"

	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"
	cliimgv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchDockerImage fetches a docker image based on a ImageStreamTag with the defined key (namespace and name).
// Returns nil if not found
func FetchDockerImage(cli cliimgv1.ImageV1Interface, key client.ObjectKey) (*dockerv10.DockerImage, error) {
	dockerImage := &dockerv10.DockerImage{}
	isTag, err := FecthImageStreamTag(cli, key, "")
	if err != nil {
		return nil, err
	} else if isTag == nil {
		return nil, nil
	}
	log.Debugf("Found image '%s' in the namespace '%s'", key.Name, key.Namespace)
	// is there any metadata to read from?
	if len(isTag.Image.DockerImageMetadata.Raw) != 0 {
		err = json.Unmarshal(isTag.Image.DockerImageMetadata.Raw, dockerImage)
		if err != nil {
			return nil, err
		}
		return dockerImage, nil
	}

	log.Warnf("Can't find any metadata in the docker image for the imagestream '%s' in the namespace '%s'", key.Name, key.Namespace)
	return nil, nil
}

// FecthImageStreamTag fetches for a particular imagestreamtag on OpenShift cluster.
// If tag is nil or empty, will search for "latest".
// Returns nil if the object was not found.
func FecthImageStreamTag(cli cliimgv1.ImageV1Interface, key client.ObjectKey, tag string) (*imgv1.ImageStreamTag, error) {
	if len(tag) == 0 {
		tag = definitions.ImageTagLatest
	}
	tagRefName := fmt.Sprintf("%s:%s", key.Name, tag)
	isTag, err := cli.ImageStreamTags(key.Namespace).Get(tagRefName, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		log.Debugf("Image '%s' not found", tagRefName)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return isTag, err
}

// CreateImageStreamTagIfNotExists will create a new ImageStreamTag if not exists
func CreateImageStreamTagIfNotExists(cli cliimgv1.ImageV1Interface, is *imgv1.ImageStreamTag) (bool, error) {
	is, err := cli.ImageStreamTags(is.Namespace).Create(is)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Debugf("Error while creating Image Stream Tag '%s' in namespace '%s'", is.Name, is.Namespace)
		return false, err
	} else if errors.IsAlreadyExists(err) {
		log.Debug("Image Stream Tag already exists in the namespace")
		return false, nil
	}
	log.Debugf("Image Stream Tag %s created in namespace %s", is.Name, is.Namespace)
	return true, nil
}
