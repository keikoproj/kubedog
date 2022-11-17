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

//Package aws provides steps implementations related to AWS.
package aws

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/keikoproj/kubedog/pkg/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	ASClient         autoscalingiface.AutoScalingAPI
	AsgName          string
	LaunchConfigName string
}

var (
	ClusterAWSRegion = common.GetEnv("AWS_REGION", "us-west-2")
	BDDClusterName   = common.GetEnv("CLUSTER_NAME", common.GetUsernamePrefix()+"kubedog-bdd")
)

/*
AnASGNamed updates the current ASG to be used by the other ASG related steps.
*/
func (c *Client) AnASGNamed(name string) error {
	if c.ASClient == nil {
		return errors.Errorf("Unable to get ASG %v: The AS client was not found, use the method GetAWSCredsAndClients", name)
	}

	out, err := c.ASClient.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
	})
	if err != nil {
		return errors.Errorf("Failed describing the ASG %v: %v", name, err)
	} else if len(out.AutoScalingGroups) == 0 {
		return errors.Errorf("No ASG found by the name: '%s'", name)
	}

	arn := aws.StringValue(out.AutoScalingGroups[0].AutoScalingGroupARN)
	log.Infof("[KUBEDOG] Auto Scaling group: %v", arn)

	c.LaunchConfigName = aws.StringValue(out.AutoScalingGroups[0].LaunchConfigurationName)
	c.AsgName = name

	return nil
}

/*
ScaleCurrentASG scales the max and min size of the current ASG.
*/
func (c *Client) ScaleCurrentASG(desiredMin, desiredMax int64) error {

	if c.ASClient == nil {
		return errors.Errorf("Unable to scale currrent ASG: The AS client was not found, use the method GetAWSCredsAndClients")
	}

	_, err := c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(c.AsgName),
		MinSize:              aws.Int64(desiredMin),
		MaxSize:              aws.Int64(desiredMax),
	})
	if err != nil {
		return errors.Errorf("Failed scaling ASG %v: %v", c.AsgName, err)
	}

	return nil
}

/*
UpdateFieldOfCurrentASG updates the current ASG. Fields/parameters supported: LaunchConfigurationName, MinSize, DesiredCapacity and MaxSize.
*/
func (c *Client) UpdateFieldOfCurrentASG(field, value string) error {
	var (
		err        error
		valueInt64 int64
	)

	if c.ASClient == nil {
		return errors.Errorf("Unable to update current ASG: The AS client was not found, use the method GetAWSCredsAndClients")
	}

	if field == "LaunchConfigurationName" {
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName:    aws.String(c.AsgName),
			LaunchConfigurationName: aws.String(value),
		})

		if err != nil {
			return errors.Errorf("Failed updating field %v to %v of ASG %v: %v", field, value, c.AsgName, err)
		}
		return nil
	}

	valueInt64, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return errors.Errorf("Failed to convert %v to int64 while trying to update field %v of ASG %v", value, field, c.AsgName)
	}

	switch field {
	case "MinSize":
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(c.AsgName),
			MinSize:              aws.Int64(valueInt64),
		})
	case "DesiredCapacity":
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(c.AsgName),
			DesiredCapacity:      aws.Int64(valueInt64),
		})
	case "MaxSize":
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(c.AsgName),
			MaxSize:              aws.Int64(valueInt64),
		})
	default:
		return errors.Errorf("Field %v is not supported, use LaunchConfigurationName, MinSize, DesiredCapacity or MaxSize", field)
	}

	if err != nil {
		return errors.Errorf("Failed updating field %v to %v of ASG %v: %v", field, value, c.AsgName, err)
	}
	return nil
}

/*
GetAWSCredsAndClients checks if there is a valid credential available and uses it to update the AS Client.
*/
func (c *Client) GetAWSCredsAndClients() error {
	var (
		sess     *session.Session
		identity *sts.GetCallerIdentityOutput
		err      error
	)

	if sess, err = session.NewSession(); err != nil {
		return err
	}

	svc := sts.New(sess)

	if identity, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{}); err != nil {
		return err
	}

	arn := aws.StringValue(identity.Arn)
	log.Infof("[KUBEDOG] Credentials: %v", arn)

	c.ASClient = autoscaling.New(sess)

	return nil
}

func (c *Client) IamRoleTrust(action, entityName, roleName string) error {
	// Load clients so we can derive account info.
	sess := GetAWSSession(ClusterAWSRegion)
	iamClient := iam.New(sess)
	stsClient := sts.New(sess)
	accountId := GetAccountNumber(stsClient)

	// Add efs-csi-role-<clustername> as trusted entity
	var trustedEntityArn = fmt.Sprintf("arn:aws:iam::%s:role/%s",
		accountId, entityName)

	type StatementEntry struct {
		Effect    string
		Action    string
		Principal map[string]string
	}
	type PolicyDocument struct {
		Version   string
		Statement []StatementEntry
	}
	newPolicyDoc := &PolicyDocument{
		Version:   "2012-10-17",
		Statement: make([]StatementEntry, 0),
	}

	role, err := GetIamRole(roleName, iamClient)
	if err != nil {
		return err
	}

	if role.AssumeRolePolicyDocument != nil {
		doc := &PolicyDocument{}
		data, err := url.QueryUnescape(*role.AssumeRolePolicyDocument)
		if err != nil {
			return err
		}

		// parse existing policy
		err = json.Unmarshal([]byte(data), &doc)
		if err != nil {
			return err
		}

		if len(doc.Statement) > 0 {
			// loop through existing statements and keep valid trusted entities
			for _, stmnt := range doc.Statement {
				if val, ok := stmnt.Principal["AWS"]; ok {
					if strings.HasPrefix(val, "arn:aws:iam") && val != trustedEntityArn {
						newPolicyDoc.Statement = append(newPolicyDoc.Statement, stmnt)
					}
				}
			}
		}
	}

	switch action {
	case "add":
		newStatment := StatementEntry{
			Effect: "Allow",
			Principal: map[string]string{
				"AWS": trustedEntityArn,
			},
			Action: "sts:AssumeRole",
		}

		newPolicyDoc.Statement = append(newPolicyDoc.Statement, newStatment)
	case "remove":
		// Do nothing, we already cleansed the trusted entity role above if we're not adding
	}

	policyJSON, err := json.Marshal(newPolicyDoc)
	if err != nil {
		return err
	}

	_, err = UpdateIAMAssumeRole(roleName, policyJSON, iamClient)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) EfsCsiRoleTrust(action, roleName string) error {
	var trustedEntityName = fmt.Sprintf("efs-csi-role-%s", BDDClusterName)
	return c.IamRoleTrust(action, trustedEntityName, roleName)
}

func (c *Client) ClusterSharedIamOperation(operation string) error {
	// Load clients so we can derive account info.
	var (
		sess      = GetAWSSession(ClusterAWSRegion)
		iamClient = iam.New(sess)
		stsClient = sts.New(sess)
		accountId = GetAccountNumber(stsClient)
		roleName  = fmt.Sprintf("shared.%s", BDDClusterName)
		iamFmt    = "arn:aws:iam::%s:%s/%s"
	)

	// for trusted entity
	clusterSharedrole := fmt.Sprintf(iamFmt, accountId, "role", roleName)
	rootIAM := fmt.Sprintf("arn:aws:iam::%s:%s", accountId, "root")
	clusterSharedPolicy := fmt.Sprintf(iamFmt, accountId, "policy", roleName)
	assumeRoleDoc := `{"Version":"2012-10-17","Statement":[{"Effect": "Allow", "Action": "sts:AssumeRole", "Principal": {"AWS": "%s"}}]}`
	roleDocument := []byte(fmt.Sprintf(assumeRoleDoc, rootIAM))
	policyDocT := `{"Version":"2012-10-17","Statement":[{"Effect": "Allow", "Action": "sts:AssumeRole", "Resource": "%s"}]}`
	policyDocument := []byte(fmt.Sprintf(policyDocT, clusterSharedrole))

	switch operation {
	case "add":
		role, err := PutIAMRole(roleName, "shared cluster role", roleDocument, iamClient)
		if err != nil {
			return errors.Wrap(err, "failed to create shared cluster role")
		}
		log.Infof("BDD >> created shared iam role: %s", awssdk.StringValue(role.Arn))

		policy, err := PutManagedPolicy(roleName, clusterSharedPolicy, "shared cluster policy", policyDocument, iamClient)
		if err != nil {
			return errors.Wrap(err, "failed to create shared cluster managed policy")
		}
		log.Infof("BDD >> created shared iam policy: %s", awssdk.StringValue(policy.Arn))
	case "remove":
		err := DeleteManagedPolicy(clusterSharedPolicy, iamClient)
		if err != nil {
			return errors.Wrap(err, "failed to delete shared cluster role")
		}

		err = DeleteIAMRole(roleName, iamClient)
		if err != nil {
			return errors.Wrap(err, "failed to delete shared cluster managed policy")
		}
	}
	return nil
}
