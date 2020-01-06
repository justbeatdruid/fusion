package names

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/docker/docker/pkg/stringid"
)

func NewID() string {
	return stringid.GenerateRandomID()
}

func NewName(i int) (string, error) {
	name := namesgenerator.GetRandomName(i)
	// name maybe a service name
	// a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character
	// see k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/validation/validation.go#IsDNS1035Label
	name = strings.Replace(name, "_", "-", -1)
	return name, nil
}

func TruncateID(id string) string {
	return stringid.TruncateID(id)
}

func TimedName(name string) string {
	return fmt.Sprintf("%s-at-%d", name, time.Now().Unix())
}
