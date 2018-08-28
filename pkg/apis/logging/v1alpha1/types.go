package v1alpha1

import (
	"bytes"
	"fmt"
	"github.com/banzaicloud/logging-operator/pkg/plugins"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"text/template"
)

// LoggingOperatorList auto generated by the sdk
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LoggingOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []LoggingOperator `json:"items"`
}

// LoggingOperator auto generated by the sdk
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LoggingOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              LoggingOperatorSpec   `json:"spec"`
	Status            LoggingOperatorStatus `json:"status,omitempty"`
}

// LoggingOperatorSpec holds the spec for the operator
type LoggingOperatorSpec struct {
	Input  Input    `json:"input"`
	Filter []Plugin `json:"filter"`
	Output []Plugin `json:"output"`
}

// LoggingOperatorStatus holds the status info for the operator
type LoggingOperatorStatus struct {
	// Fill me
}

// Input this determines the log origin
type Input struct {
	Label map[string]string `json:"label"`
}

// Plugin struct for fluentd plugins
type Plugin struct {
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Parameters []Parameter `json:"parameters"`
}

// RenderPlugin general Plugin renderer
func RenderPlugin(plugin Plugin, baseMap map[string]string) (string, error) {
	rawTemplate, err := plugins.GetTemplate(plugin.Type)
	if err != nil {
		return "", err
	}
	for _, param := range plugin.Parameters {
		k, v := param.GetValue()
		baseMap[k] = v
	}

	t := template.New("PluginTemplate")
	t, err = t.Parse(rawTemplate)
	if err != nil {
		return "", err
	}
	tpl := new(bytes.Buffer)
	err = t.Execute(tpl, baseMap)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// Parameter generic parameter type to handle values from different sources
type Parameter struct {
	Name      string     `json:"name"`
	ValueFrom *ValueFrom `json:"valueFrom"`
	Value     string     `json:"value"`
}

// GetValue for a Parameter
func (p Parameter) GetValue() (string, string) {
	if p.ValueFrom != nil {
		value, error := p.ValueFrom.GetValue()
		if error != nil {
			logrus.Error(error)
			return "", ""
		}
		return p.Name, value
	}
	return p.Name, p.Value
}

// ValueFrom generic type to determine value origin
type ValueFrom struct {
	SecretKeyRef KubernetesSecret `json:"secretKeyRef"`
}

// GetValue handles the different origin of ValueFrom
func (vf *ValueFrom) GetValue() (string, error) {
	return vf.SecretKeyRef.GetValue()
}

// KubernetesSecret is a ValueFrom type
type KubernetesSecret struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// GetValue implement GetValue interface
func (ks KubernetesSecret) GetValue() (string, error) {
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      ks.Name,
		},
	}
	err := sdk.Get(&secret)
	if err != nil {
		return "", err
	}
	value, ok := secret.Data[ks.Key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %q ", ks.Key, ks.Name)
	}
	return string(value), nil
}