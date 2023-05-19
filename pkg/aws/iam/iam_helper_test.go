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

package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

func TestCreateIAMRole(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	iamClient := &FakeIAMClient{}
	policyYAML := `|
  Version: '2012-10-17'
  Statement:
  - Effect: Allow
    Action:
    - autoscaling:TerminateInstanceInAutoScalingGroup
    - autoscaling:DescribeAutoScalingGroups
    - ec2:DescribeTags
    - ec2:DescribeInstances
    Resource:
    - "*"`
	policyJSON, err := yaml.YAMLToJSON([]byte(policyYAML))
	g.Expect(err).To(gomega.BeNil())

	output, err := createIAMRole("arn:aws:iam::aws:policy/test-role", "Description", policyJSON, iamClient)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(output).ToNot(gomega.BeNil())
}

func TestDeleteManagedPolicyVersion(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	iamClient := &FakeIAMClient{}

	err := deleteManagedPolicyVersion("arn:aws:iam::aws:policy/test-role", "testid", iamClient)
	g.Expect(err).To(gomega.BeNil())
}

func TestCreateManagedPolicy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	iamClient := &FakeIAMClient{}
	policyYAML := `|
  Version: '2012-10-17'
  Statement:
  - Effect: Allow
    Action:
    - autoscaling:TerminateInstanceInAutoScalingGroup
    - autoscaling:DescribeAutoScalingGroups
    - ec2:DescribeTags
    - ec2:DescribeInstances
    Resource:
    - "*"`
	policyJSON, err := yaml.YAMLToJSON([]byte(policyYAML))
	g.Expect(err).To(gomega.BeNil())

	_, err = createManagedPolicy("arn:aws:iam::aws:policy/test-role", "Description", policyJSON, iamClient)
	g.Expect(err).To(gomega.BeNil())
}

func TestGetOldestVersionID(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	var versions []*iam.PolicyVersion
	versions = append(versions, &iam.PolicyVersion{VersionId: aws.String("v1"), CreateDate: aws.Time(time.Now().Add(5 * time.Minute))})
	versions = append(versions, &iam.PolicyVersion{VersionId: aws.String("v2"), CreateDate: aws.Time(time.Now())})
	versions = append(versions, &iam.PolicyVersion{VersionId: aws.String("v3"), CreateDate: aws.Time(time.Now().Add(10 * time.Minute))})
	versions = append(versions, &iam.PolicyVersion{VersionId: aws.String("v4"), CreateDate: aws.Time(time.Now())})

	oldestId := getOldestVersionID(versions)
	g.Expect(oldestId).To(gomega.Equal("v2"))
}
