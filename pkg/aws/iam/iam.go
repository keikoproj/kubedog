package iam

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	util "github.com/keikoproj/kubedog/internal/utilities"
	log "github.com/sirupsen/logrus"
)

func GetIamRole(roleName string, iamClient iamiface.IAMAPI) (*iam.Role, error) {
	params := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	out, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.GetRole(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get iam role %q. %v", roleName, err)
	}

	return out.(*iam.GetRoleOutput).Role, nil
}

func PutIAMRole(name, description string, policyJSON []byte, iamClient iamiface.IAMAPI, tags ...*iam.Tag) (*iam.Role, error) {
	out, err := GetIamRole(name, iamClient)
	if err != nil {
		out, err := createIAMRole(name, description, policyJSON, iamClient, tags...)
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	// If role already exits just update assume role policy
	_, err = UpdateIAMAssumeRole(name, policyJSON, iamClient)
	if err != nil {
		return out, err
	}
	return out, nil
}

func UpdateIAMAssumeRole(roleName string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.UpdateAssumeRolePolicyOutput, error) {
	json := string(policyJSON)
	params := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(json),
	}
	out, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.UpdateAssumeRolePolicy(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update assume role policy for %q .%v", roleName, err)
	}

	return out.(*iam.UpdateAssumeRolePolicyOutput), nil
}

func PutManagedPolicy(name, arn, description string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.Policy, error) {
	existingPolicy, existingPolicyVersion, err := getManagedPolicy(arn, iamClient)
	if err != nil {
		if strings.Contains(err.Error(), iam.ErrCodeNoSuchEntityException) {
			out, err := createManagedPolicy(name, description, policyJSON, iamClient)
			if err != nil {
				return nil, fmt.Errorf("failed to create managed policy %q. %v", name, err)
			}
			return out, nil
		}
		return nil, err
	}

	existingPolicyDoc, err := url.QueryUnescape(*existingPolicyVersion.Document)
	if err != nil {
		return nil, fmt.Errorf("failed to url decode managed policy document %q. %v", name, err)
	}
	if reflect.DeepEqual(existingPolicyDoc, string(policyJSON)) {
		log.Infof("managed policy %s is same as before, hence not updating", name)
		return existingPolicy, nil
	}

	out, err := updateManagedPolicy(arn, policyJSON, iamClient)
	if err != nil {
		return nil, fmt.Errorf("failed to update managed policy %q. %v", arn, err)
	}

	return &iam.Policy{
		Arn:                           existingPolicy.Arn,
		CreateDate:                    out.CreateDate,
		DefaultVersionId:              out.VersionId,
		PermissionsBoundaryUsageCount: existingPolicy.PermissionsBoundaryUsageCount,
		PolicyId:                      existingPolicy.PolicyId,
		PolicyName:                    existingPolicy.PolicyName,
	}, nil
}

func DeleteManagedPolicy(arn string, iamClient iamiface.IAMAPI) error {
	// first list all version in order to delete them
	policyVersions, err := listManagedPolicyVersions(arn, iamClient)
	if err != nil {
		return fmt.Errorf("failed to list managed policy versions %q. %v", arn, err)
	}

	// delete all versions except default, as default version can only be deleted with policy
	for _, ver := range policyVersions {
		if !(*ver.IsDefaultVersion) {
			err := deleteManagedPolicyVersion(arn, *ver.VersionId, iamClient)
			if err != nil {
				return fmt.Errorf("failed to delete managed policy %q version %q. %v", arn, *ver.VersionId, err)
			}
		}
	}

	// now delete policy
	params := &iam.DeletePolicyInput{
		PolicyArn: aws.String(arn),
	}
	_, err = util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.DeletePolicy(params)
	})
	if err != nil {
		return err
	}
	return nil
}

func DeleteIAMRole(roleName string, iamClient iamiface.IAMAPI) error {
	params := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}
	_, err := util.RetryOnError(&util.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.DeleteRole(params)
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "NoSuchEntity" {
				return nil
			}
		}
		return fmt.Errorf("failed to delete iam role %q. %v", roleName, err)
	}

	return nil
}
