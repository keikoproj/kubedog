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
	"github.com/onsi/gomega"
)

type mockAutoScalingClient struct {
	autoscalingiface.AutoScalingAPI
	ASGs []*autoscaling.Group
	Err  error
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
			{ // case len(out.AutoScalingGroups) > 1
				ASClient: mockAutoScalingClient{
					ASGs: []*autoscaling.Group{
						{
							AutoScalingGroupName:    aws.String("ASG-name-1"),
							LaunchConfigurationName: aws.String("ASG-name-1-LC"),
							AutoScalingGroupARN:     aws.String("ASG-name-1-ARN"),
						},
						{
							AutoScalingGroupName:    aws.String("ASG-name-1"),
							LaunchConfigurationName: aws.String("ASG-name-2-LC"),
							AutoScalingGroupARN:     aws.String("ASG-name-3-ARN"),
						},
					},
					Err: nil,
				},
				expectedASG: &autoscaling.Group{
					AutoScalingGroupName: aws.String("ASG-name-1"),
				},
				expectError: true,
			},
		}
	)

	// Not ASClient
	ASC := Client{}
	err := ASC.AnASGNamed("Some-ASG-Name")
	g.Expect(err).Should(gomega.HaveOccurred())

	for _, test := range tests {
		client := Client{ASClient: &test.ASClient}
		ASG_Name := aws.StringValue(test.expectedASG.AutoScalingGroupName)
		err := client.AnASGNamed(ASG_Name)

		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
			g.Expect(client.AsgName).To(gomega.Equal(""))
			g.Expect(client.LaunchConfigName).To(gomega.Equal(""))
		} else {
			LC_Name := aws.StringValue(test.expectedASG.LaunchConfigurationName)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			g.Expect(client.AsgName).To(gomega.Equal(ASG_Name))
			g.Expect(client.LaunchConfigName).To(gomega.Equal(LC_Name))
		}
	}
}
func TestPositiveUpdateFieldOfCurrentASG(t *testing.T) {
	var (
		g   = gomega.NewWithT(t)
		ASC = Client{
			ASClient: &mockAutoScalingClient{
				Err: nil,
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
				Err: errors.New("some-UASG-error"),
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
}

func TestPositiveScaleCurrentASG(t *testing.T) {
	var (
		g   = gomega.NewWithT(t)
		ASC = Client{
			ASClient: &mockAutoScalingClient{
				Err: nil,
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
				Err: errors.New("some-UASG-error"),
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
	return &autoscaling.UpdateAutoScalingGroupOutput{}, asc.Err
}

func (asc *mockAutoScalingClient) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	ASGs := []*autoscaling.Group{}
	for _, inName := range aws.StringValueSlice(input.AutoScalingGroupNames) {
		for _, Group := range asc.ASGs {
			if aws.StringValue(Group.AutoScalingGroupName) == inName {
				ASGs = append(ASGs, Group)
				//No break to allow case 'len(out.AutoScalingGroups) > 1' to happen
			}
		}
	}
	out := &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: ASGs,
	}
	return out, asc.Err
}
