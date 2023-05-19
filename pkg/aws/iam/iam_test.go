package iam

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

// mock IAM client
type FakeIAMClient struct {
	iamiface.IAMAPI
	RoleArn string
}

func (fiam *FakeIAMClient) CreateRole(roleInput *iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	output := &iam.CreateRoleOutput{
		Role: &iam.Role{RoleName: roleInput.RoleName},
	}
	if fiam.RoleArn != "" {
		output.Role.Arn = aws.String(fiam.RoleArn)
	}
	return output, nil
}

func (fiam *FakeIAMClient) GetRole(roleInput *iam.GetRoleInput) (*iam.GetRoleOutput, error) {
	output := &iam.GetRoleOutput{
		Role: &iam.Role{RoleName: roleInput.RoleName},
	}
	if fiam.RoleArn != "" {
		output.Role.Arn = aws.String(fiam.RoleArn)
	}
	return output, nil
}

func (fiam *FakeIAMClient) DeleteRole(delRoleInput *iam.DeleteRoleInput) (*iam.DeleteRoleOutput, error) {
	return &iam.DeleteRoleOutput{}, nil
}

func (fiam *FakeIAMClient) UpdateAssumeRolePolicy(*iam.UpdateAssumeRolePolicyInput) (*iam.UpdateAssumeRolePolicyOutput, error) {
	return &iam.UpdateAssumeRolePolicyOutput{}, nil
}

func (fiam *FakeIAMClient) GetPolicy(*iam.GetPolicyInput) (*iam.GetPolicyOutput, error) {
	return &iam.GetPolicyOutput{
		Policy: &iam.Policy{DefaultVersionId: aws.String("v1")},
	}, nil
}

func (fiam *FakeIAMClient) GetPolicyVersion(*iam.GetPolicyVersionInput) (*iam.GetPolicyVersionOutput, error) {
	return &iam.GetPolicyVersionOutput{
		PolicyVersion: &iam.PolicyVersion{Document: aws.String("Version: '2012-10-17'")},
	}, nil
}

func (fiam *FakeIAMClient) CreatePolicy(*iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error) {
	return &iam.CreatePolicyOutput{}, nil
}

func (fiam *FakeIAMClient) ListPolicyVersions(*iam.ListPolicyVersionsInput) (*iam.ListPolicyVersionsOutput, error) {
	return &iam.ListPolicyVersionsOutput{}, nil
}

func (fiam *FakeIAMClient) CreatePolicyVersion(*iam.CreatePolicyVersionInput) (*iam.CreatePolicyVersionOutput, error) {
	return &iam.CreatePolicyVersionOutput{
		PolicyVersion: &iam.PolicyVersion{
			CreateDate:       aws.Time(time.Now()),
			IsDefaultVersion: aws.Bool(true),
			VersionId:        aws.String("v2"),
		},
	}, nil
}

func (fiam *FakeIAMClient) DeletePolicy(*iam.DeletePolicyInput) (*iam.DeletePolicyOutput, error) {
	return &iam.DeletePolicyOutput{}, nil
}

func (fiam *FakeIAMClient) DeletePolicyVersion(input *iam.DeletePolicyVersionInput) (*iam.DeletePolicyVersionOutput, error) {
	return &iam.DeletePolicyVersionOutput{}, nil
}

func TestDeleteIAMRole(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	iamClient := &FakeIAMClient{}

	err := DeleteIAMRole("arn:aws:iam::aws:policy/test-role", iamClient)
	g.Expect(err).To(gomega.BeNil())
}

func TestGetIamRole(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	iamClient := &FakeIAMClient{}

	output, err := GetIamRole("arn:aws:iam::aws:policy/test-role", iamClient)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(output).ToNot(gomega.BeNil())
}

func TestUpdateIAMAssumeRole(t *testing.T) {
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

	output, err := UpdateIAMAssumeRole("arn:aws:iam::aws:policy/test-role", policyJSON, iamClient)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(output).ToNot(gomega.BeNil())
}

func TestPutIAMRole(t *testing.T) {
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

	output, err := PutIAMRole("arn:aws:iam::aws:policy/test-role", "Description", policyJSON, iamClient)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(output).ToNot(gomega.BeNil())
}

func TestPutManagedPolicy(t *testing.T) {
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

	output, err := PutManagedPolicy("arn:aws:iam::aws:policy/test-role", "arn:aws:iam::aws:policy/test-arn", "Description", policyJSON, iamClient)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(output).ToNot(gomega.BeNil())
}

func TestDeleteManagedPolicy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	iamClient := &FakeIAMClient{}

	err := DeleteManagedPolicy("arn:aws:iam::aws:policy/test-role", iamClient)
	g.Expect(err).To(gomega.BeNil())
}
