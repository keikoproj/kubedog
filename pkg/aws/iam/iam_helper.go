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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	util "github.com/keikoproj/kubedog/internal/utilities"
)

func getManagedPolicy(policyARN string, iamClient iamiface.IAMAPI) (*iam.Policy, *iam.PolicyVersion, error) {
	policyParams := &iam.GetPolicyInput{
		PolicyArn: aws.String(policyARN),
	}
	out, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.GetPolicy(policyParams)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get managed policy %q. %v", policyARN, err)
	}
	policyVersionParams := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(policyARN),
		VersionId: out.(*iam.GetPolicyOutput).Policy.DefaultVersionId,
	}
	policyVersionOut, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.GetPolicyVersion(policyVersionParams)
	})
	if err != nil {
		return out.(*iam.GetPolicyOutput).Policy, nil, fmt.Errorf("failed to get managed policy version %q. %v", policyARN, err)
	}
	return out.(*iam.GetPolicyOutput).Policy, policyVersionOut.(*iam.GetPolicyVersionOutput).PolicyVersion, nil
}

func updateManagedPolicy(arn string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.PolicyVersion, error) {
	versions, err := listManagedPolicyVersions(arn, iamClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed policy versions %q. %v", arn, err)
	}

	if len(versions) >= 3 {
		verId := getOldestVersionID(versions)
		err := deleteManagedPolicyVersion(arn, verId, iamClient)
		if err != nil {
			return nil, fmt.Errorf("failed to delete managed policy %q version %q. %v", arn, verId, err)
		}
	}

	out, err := createManagedPolicyVersion(arn, policyJSON, true, iamClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed policy version %q. %v", arn, err)
	}

	return out, nil
}

func createManagedPolicyVersion(arn string, policyJSON []byte, isDefault bool, iamClient iamiface.IAMAPI) (*iam.PolicyVersion, error) {
	json := string(policyJSON)
	params := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(arn),
		PolicyDocument: aws.String(json),
		SetAsDefault:   aws.Bool(isDefault),
	}
	out, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.CreatePolicyVersion(params)
	})
	if err != nil {
		return nil, fmt.Errorf("faild to create managed policy version %q. %v", arn, err)
	}
	return out.(*iam.CreatePolicyVersionOutput).PolicyVersion, nil
}

func createManagedPolicy(name, description string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.Policy, error) {
	json := string(policyJSON)
	params := &iam.CreatePolicyInput{
		Description:    aws.String(description),
		PolicyDocument: aws.String(json),
		PolicyName:     aws.String(name),
	}

	out, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.CreatePolicy(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create managed policy %q. %v", name, err)
	}

	return out.(*iam.CreatePolicyOutput).Policy, nil
}

func listManagedPolicyVersions(arn string, iamClient iamiface.IAMAPI) ([]*iam.PolicyVersion, error) {
	params := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}
	listVersionsOutput, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.ListPolicyVersions(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list managed policy versions %q. %v", arn, err)
	}
	return listVersionsOutput.(*iam.ListPolicyVersionsOutput).Versions, nil
}

func getOldestVersionID(versions []*iam.PolicyVersion) string {
	earliestTime := *versions[0].CreateDate
	oldestVersionID := *versions[0].VersionId
	for _, ver := range versions {
		if earliestTime.After(*ver.CreateDate) {
			oldestVersionID = *ver.VersionId
			earliestTime = *ver.CreateDate
		}
	}
	return oldestVersionID
}

func deleteManagedPolicyVersion(arn, id string, iamClient iamiface.IAMAPI) error {
	params := &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(id),
	}
	_, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.DeletePolicyVersion(params)
	})
	if err != nil {
		return fmt.Errorf("failed to delete managed policy %q version %q. %v", arn, id, err)
	}
	return nil
}

func isThrottling(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "Throttling" {
			return true
		}
	}
	return false
}

func createIAMRole(name, description string, policyJSON []byte, iamClient iamiface.IAMAPI, tags ...*iam.Tag) (*iam.Role, error) {
	json := string(policyJSON)
	role := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(json),
		Description:              aws.String(description),
	}
	if len(tags) > 0 {
		role.Tags = tags
	}
	out, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.CreateRole(role)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create iam role with policy %q. %v", name, err)
	}

	return out.(*iam.CreateRoleOutput).Role, nil
}
