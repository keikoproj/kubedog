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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/onsi/gomega"
)

const (
	TestAwsAccountNumber = "0000123456789"
)

type mockAutoScalingClient struct {
	autoscalingiface.AutoScalingAPI
	ASGs []*autoscaling.Group
	Err  error
}

type STSMocker struct {
	stsiface.STSAPI
}

func (s *STSMocker) GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	output := &sts.GetCallerIdentityOutput{
		Account: aws.String(TestAwsAccountNumber),
	}

	return output, nil
}

func TestAnASGNamed(t *testing.T) {
	var (
		g     = gomega.NewWithT(t)
		tests = []struct {
			ASClient    mockAutoScalingClient
			expectedASG *autoscaling.Group
			expectError bool
		}{
			{ // ASClient.DescribeAutoScalingGroups fails
				ASClient: mockAutoScalingClient{
					Err: errors.New("some DescribeAutoScalingGroups error"),
				},
				expectedASG: &autoscaling.Group{
					AutoScalingGroupName: aws.String("some-ASG-name"),
				},
				expectError: true,
			},
			{ // case len(out.AutoScalingGroups) = 0
				ASClient: mockAutoScalingClient{
					ASGs: []*autoscaling.Group{
						{
							AutoScalingGroupName:    aws.String("ASG-name-1"),
							LaunchConfigurationName: aws.String("ASG-name-1-LC"),
							AutoScalingGroupARN:     aws.String("ASG-name-1-ARN"),
						},
					},
					Err: nil,
				},
				expectedASG: &autoscaling.Group{
					AutoScalingGroupName: aws.String("ASG-name-2"),
				},
				expectError: true,
			},
			{ // case len(out.AutoScalingGroups) = 1
				ASClient: mockAutoScalingClient{
					ASGs: []*autoscaling.Group{
						{
							AutoScalingGroupName:    aws.String("ASG-name-1"),
							LaunchConfigurationName: aws.String("ASG-name-1-LC"),
							AutoScalingGroupARN:     aws.String("ASG-name-1-ARN"),
						},
						{
							AutoScalingGroupName:    aws.String("ASG-name-2"),
							LaunchConfigurationName: aws.String("ASG-name-2-LC"),
							AutoScalingGroupARN:     aws.String("ASG-name-2-ARN"),
						},
					},
					Err: nil,
				},
				expectedASG: &autoscaling.Group{
					AutoScalingGroupName:    aws.String("ASG-name-2"),
					LaunchConfigurationName: aws.String("ASG-name-2-LC"),
				},
				expectError: false,
			},
		}
	)

	// Not ASClient
	ASC := ClientSet{}
	err := ASC.AnASGNamed("Some-ASG-Name")
	g.Expect(err).Should(gomega.HaveOccurred())

	for _, test := range tests {
		client := ClientSet{ASClient: &test.ASClient}
		ASG_Name := aws.StringValue(test.expectedASG.AutoScalingGroupName)
		err := client.AnASGNamed(ASG_Name)

		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
			g.Expect(client.asgName).To(gomega.Equal(""))
			g.Expect(client.launchConfigName).To(gomega.Equal(""))
		} else {
			LC_Name := aws.StringValue(test.expectedASG.LaunchConfigurationName)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			g.Expect(client.asgName).To(gomega.Equal(ASG_Name))
			g.Expect(client.launchConfigName).To(gomega.Equal(LC_Name))
		}
	}
}
func TestPositiveUpdateFieldOfCurrentASG(t *testing.T) {
	var (
		g   = gomega.NewWithT(t)
		ASC = ClientSet{
			ASClient: &mockAutoScalingClient{
				Err: nil,
			},
			asgName:          "asg-test",
			launchConfigName: "current-lc-asg-test",
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
		emptyASC = ClientSet{}
		ASC      = ClientSet{
			ASClient: &mockAutoScalingClient{
				Err: errors.New("some-UASG-error"),
			},
			asgName:          "asg-test",
			launchConfigName: "current-lc-asg-test",
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
}

func TestPositiveScaleCurrentASG(t *testing.T) {
	var (
		g   = gomega.NewWithT(t)
		ASC = ClientSet{
			ASClient: &mockAutoScalingClient{
				Err: nil,
			},
			asgName: "asg-test",
		}
	)

	const someNumber int64 = 3

	err := ASC.ScaleCurrentASG(someNumber, someNumber)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestNegativeScaleCurrentASG(t *testing.T) {
	var (
		g        = gomega.NewWithT(t)
		emptyASC = ClientSet{}
		ASC      = ClientSet{
			ASClient: &mockAutoScalingClient{
				Err: errors.New("some-UASG-error"),
			},
			asgName: "asg-test",
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
	return &autoscaling.UpdateAutoScalingGroupOutput{}, asc.Err
}

func (asc *mockAutoScalingClient) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	ASGs := []*autoscaling.Group{}
	for _, inName := range aws.StringValueSlice(input.AutoScalingGroupNames) {
		for _, Group := range asc.ASGs {
			if aws.StringValue(Group.AutoScalingGroupName) == inName {
				ASGs = append(ASGs, Group)
				break
			}
		}
	}
	out := &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: ASGs,
	}
	return out, asc.Err
}
