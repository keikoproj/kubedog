/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/onsi/gomega"
)

type mockAutoScalingClient struct {
	autoscalingiface.AutoScalingAPI
	UpdateAutoScalingGroupErr error
}

func TestPositiveUpdateFieldOfCurrentASG(t *testing.T) {
	var (
		g   = gomega.NewWithT(t)
		ASC = Client{
			ASClient: &mockAutoScalingClient{
				UpdateAutoScalingGroupErr: nil,
			},
			AsgName:          "asg-test",
			LaunchConfigName: "current-lc-asg-test",
		}
	)

	const (
		someLaunchConfigName = "new-lc-asg-test"
		someNumber           = "3"
	)

	err := ASC.UpdateFieldOfCurrentASG("LaunchConfigurationName", someLaunchConfigName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = ASC.UpdateFieldOfCurrentASG("MinSize", someNumber)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = ASC.UpdateFieldOfCurrentASG("DesiredCapacity", someNumber)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = ASC.UpdateFieldOfCurrentASG("MaxSize", someNumber)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestNegativeUpdateFieldOfCurrentASG(t *testing.T) {
	var (
		g        = gomega.NewWithT(t)
		emptyASC = Client{}
		ASC      = Client{
			ASClient: &mockAutoScalingClient{
				UpdateAutoScalingGroupErr: errors.New("some-UASG-error"),
			},
			AsgName:          "asg-test",
			LaunchConfigName: "current-lc-asg-test",
		}
	)

	const (
		someLaunchConfigName = "new-lc-asg-test"
		someNumber           = "3"
	)

	// Empty client
	err := emptyASC.UpdateFieldOfCurrentASG("someField", "someValue")
	g.Expect(err).Should(gomega.HaveOccurred())
	// Invalid size value
	err = ASC.UpdateFieldOfCurrentASG("someSizeField", "someInvalidValue")
	g.Expect(err).Should(gomega.HaveOccurred())
	// Unsupported field
	err = ASC.UpdateFieldOfCurrentASG("someNotSupportedField", someNumber)
	g.Expect(err).Should(gomega.HaveOccurred())
	// Error updating Launch Config
	err = ASC.UpdateFieldOfCurrentASG("LaunchConfigurationName", someLaunchConfigName)
	g.Expect(err).Should(gomega.HaveOccurred())
	// Error updating the size
	err = ASC.UpdateFieldOfCurrentASG("MinSize", someNumber)
	g.Expect(err).Should(gomega.HaveOccurred())
	err = ASC.UpdateFieldOfCurrentASG("DesiredCapacity", someNumber)
	g.Expect(err).Should(gomega.HaveOccurred())
	err = ASC.UpdateFieldOfCurrentASG("MaxSize", someNumber)
	g.Expect(err).Should(gomega.HaveOccurred())
	//REMOVE - BREAKING TO TEST CI
	g.Expect(false).Should(gomega.BeTrue())
}

func TestPositiveScaleCurrentASG(t *testing.T) {
	var (
		g   = gomega.NewWithT(t)
		ASC = Client{
			ASClient: &mockAutoScalingClient{
				UpdateAutoScalingGroupErr: nil,
			},
			AsgName: "asg-test",
		}
	)

	const someNumber int64 = 3

	err := ASC.ScaleCurrentASG(someNumber, someNumber)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestNegativeScaleCurrentASG(t *testing.T) {
	var (
		g        = gomega.NewWithT(t)
		emptyASC = Client{}
		ASC      = Client{
			ASClient: &mockAutoScalingClient{
				UpdateAutoScalingGroupErr: errors.New("some-UASG-error"),
			},
			AsgName: "asg-test",
		}
	)

	const someNumber int64 = 3

	// Empty client
	err := emptyASC.ScaleCurrentASG(someNumber, someNumber)
	g.Expect(err).Should(gomega.HaveOccurred())

	// Error scaling ASG
	err = ASC.ScaleCurrentASG(someNumber, someNumber)
	g.Expect(err).Should(gomega.HaveOccurred())
}

func (asc *mockAutoScalingClient) UpdateAutoScalingGroup(input *autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	return &autoscaling.UpdateAutoScalingGroupOutput{}, asc.UpdateAutoScalingGroupErr
}
