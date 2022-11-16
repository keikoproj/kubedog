package aws

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/keikoproj/kubedog/pkg/common"
	log "github.com/sirupsen/logrus"
)

// GetAWSSession for a given region
func GetAWSSession(region string) client.ConfigProvider {
	var config aws.Config
	config.Region = aws.String(region)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            config,
	}))

	return sess
}

// GetAccountNumber returns AWS account number
func GetAccountNumber(svc stsiface.STSAPI) string {
	// Region is defaulted to "us-west-2"

	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		log.Infof("Failed to get caller identity: %s", err.Error())
		return ""
	}

	return *result.Account
}

// GetIamRole return existing IAM role data
func GetIamRole(roleName string, iamClient iamiface.IAMAPI) (*iam.Role, error) {
	params := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	out, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.GetRole(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get iam role %q. %v", roleName, err)
	}

	return out.(*iam.GetRoleOutput).Role, nil
}

// PutIAMRole returns updates/Creates IAM Role
func PutIAMRole(name, description string, policyJSON []byte, iamClient iamiface.IAMAPI, tags ...*iam.Tag) (*iam.Role, error) {
	out, err := GetIamRole(name, iamClient)
	if err != nil {
		out, err := CreateIAMRole(name, description, policyJSON, iamClient, tags...)
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

// CreateIAMRole Wrapper to create IAM Role with policy in AWS.
func CreateIAMRole(name, description string, policyJSON []byte, iamClient iamiface.IAMAPI, tags ...*iam.Tag) (*iam.Role, error) {
	json := string(policyJSON)
	role := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(json),
		Description:              aws.String(description),
	}
	if len(tags) > 0 {
		role.Tags = tags
	}
	out, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.CreateRole(role)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create iam role with policy %q. %v", name, err)
	}

	return out.(*iam.CreateRoleOutput).Role, nil
}

// UpdateIAMAssumeRole Updates assume role policy doc
func UpdateIAMAssumeRole(roleName string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.UpdateAssumeRolePolicyOutput, error) {
	json := string(policyJSON)
	params := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(json),
	}
	out, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.UpdateAssumeRolePolicy(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update assume role policy for %q .%v", roleName, err)
	}

	return out.(*iam.UpdateAssumeRolePolicyOutput), nil
}

// isThrottling return true if API throttling exception
func isThrottling(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "Throttling" {
			return true
		}
	}
	return false
}

// PutManagedPolicy returns updates/creates Managed IAM Policy
func PutManagedPolicy(name, arn, description string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.Policy, error) {
	// check if policy already exists
	existingPolicy, existingPolicyVersion, err := GetManagedPolicy(arn, iamClient)
	if err != nil {
		if strings.Contains(err.Error(), iam.ErrCodeNoSuchEntityException) {
			// if policy doesn't exists then create new policy
			out, err := CreateManagedPolicy(name, description, policyJSON, iamClient)
			if err != nil {
				return nil, fmt.Errorf("failed to create managed policy %q. %v", name, err)
			}
			return out, nil
		}
		return nil, err
	}

	// if policy already exists then compare current policy with existing policy, if equal then do nothing
	existingPolicyDoc, err := url.QueryUnescape(*existingPolicyVersion.Document)
	if err != nil {
		return nil, fmt.Errorf("failed to url decode managed policy document %q. %v", name, err)
	}
	if reflect.DeepEqual(existingPolicyDoc, string(policyJSON)) {
		log.Infof("managed policy %s is same as before, hence not updating", name)
		return existingPolicy, nil
	}

	// if current policy and existing policy are different then update existing policy
	out, err := UpdateManagedPolicy(arn, policyJSON, iamClient)
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

// GetManagedPolicy retrieves information about the specified managed policy and its default version
func GetManagedPolicy(policyARN string, iamClient iamiface.IAMAPI) (*iam.Policy, *iam.PolicyVersion, error) {
	policyParams := &iam.GetPolicyInput{
		PolicyArn: aws.String(policyARN),
	}
	out, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.GetPolicy(policyParams)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get managed policy %q. %v", policyARN, err)
	}
	policyVersionParams := &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(policyARN),
		VersionId: out.(*iam.GetPolicyOutput).Policy.DefaultVersionId,
	}
	policyVersionOut, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.GetPolicyVersion(policyVersionParams)
	})
	if err != nil {
		return out.(*iam.GetPolicyOutput).Policy, nil, fmt.Errorf("failed to get managed policy version %q. %v", policyARN, err)
	}
	return out.(*iam.GetPolicyOutput).Policy, policyVersionOut.(*iam.GetPolicyVersionOutput).PolicyVersion, nil
}

// UpdateManagedPolicy creates new managed policy version and set it as default
// A managed policy can have maximum of 5 versions, we are setting the threshold to be 3
// 1. List managed policy versions
// 2. Check if >= 3 (maximum threshold)
// 3. Then, delete oldest version
// 4. And, create new version and set it as default
func UpdateManagedPolicy(arn string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.PolicyVersion, error) {
	// list all versions
	versions, err := ListManagedPolicyVersions(arn, iamClient)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed policy versions %q. %v", arn, err)
	}

	if len(versions) >= 3 {
		// delete oldest policy version
		verId := getOldestVersionID(versions)
		err := DeleteManagedPolicyVersion(arn, verId, iamClient)
		if err != nil {
			return nil, fmt.Errorf("failed to delete managed policy %q version %q. %v", arn, verId, err)
		}
	}

	// create new version for this policy and set it as default
	out, err := CreateManagedPolicyVersion(arn, policyJSON, true, iamClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create managed policy version %q. %v", arn, err)
	}

	return out, nil
}

func CreateManagedPolicy(name, description string, policyJSON []byte, iamClient iamiface.IAMAPI) (*iam.Policy, error) {
	json := string(policyJSON)
	params := &iam.CreatePolicyInput{
		Description:    aws.String(description),
		PolicyDocument: aws.String(json),
		PolicyName:     aws.String(name),
	}

	out, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.CreatePolicy(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create managed policy %q. %v", name, err)
	}

	return out.(*iam.CreatePolicyOutput).Policy, nil
}

// ListManagedPolicyVersions lists specified managed policy versions
func ListManagedPolicyVersions(arn string, iamClient iamiface.IAMAPI) ([]*iam.PolicyVersion, error) {
	params := &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	}
	listVersionsOutput, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.ListPolicyVersions(params)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list managed policy versions %q. %v", arn, err)
	}
	return listVersionsOutput.(*iam.ListPolicyVersionsOutput).Versions, nil
}

// returns policy's oldest version's id
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

// DeleteManagedPolicyVersion return delete managed policy version
func DeleteManagedPolicyVersion(arn, id string, iamClient iamiface.IAMAPI) error {
	params := &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(id),
	}
	_, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.DeletePolicyVersion(params)
	})
	if err != nil {
		return fmt.Errorf("failed to delete managed policy %q version %q. %v", arn, id, err)
	}
	return nil
}

// CreateManagedPolicyVersion creates managed policy version and set it as default
func CreateManagedPolicyVersion(arn string, policyJSON []byte, isDefault bool, iamClient iamiface.IAMAPI) (*iam.PolicyVersion, error) {
	json := string(policyJSON)
	params := &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(arn),
		PolicyDocument: aws.String(json),
		SetAsDefault:   aws.Bool(isDefault),
	}
	out, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.CreatePolicyVersion(params)
	})
	if err != nil {
		return nil, fmt.Errorf("faild to create managed policy version %q. %v", arn, err)
	}
	return out.(*iam.CreatePolicyVersionOutput).PolicyVersion, nil
}

// DeleteManagedPolicy returns delete managed policy
func DeleteManagedPolicy(arn string, iamClient iamiface.IAMAPI) error {
	// first list all version in order to delete them
	policyVersions, err := ListManagedPolicyVersions(arn, iamClient)
	if err != nil {
		return fmt.Errorf("failed to list managed policy versions %q. %v", arn, err)
	}

	// delete all versions except default, as default version can only be deleted with policy
	for _, ver := range policyVersions {
		if !(*ver.IsDefaultVersion) {
			err := DeleteManagedPolicyVersion(arn, *ver.VersionId, iamClient)
			if err != nil {
				return fmt.Errorf("failed to delete managed policy %q version %q. %v", arn, *ver.VersionId, err)
			}
		}
	}

	// now delete policy
	params := &iam.DeletePolicyInput{
		PolicyArn: aws.String(arn),
	}
	_, err = common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
		return iamClient.DeletePolicy(params)
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteIAMRole Wrapper to delete IAM Role with in policy only.
func DeleteIAMRole(roleName string, iamClient iamiface.IAMAPI) error {
	params := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}
	_, err := common.RetryOnError(&common.DefaultRetry, isThrottling, func() (interface{}, error) {
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
