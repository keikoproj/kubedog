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
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	kIam "github.com/keikoproj/kubedog/pkg/aws/iam"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ClientSet struct {
	ASClient         autoscalingiface.AutoScalingAPI
	EKSClient        eksiface.EKSAPI
	Route53Client    route53iface.Route53API
	IAMClient        iamiface.IAMAPI
	STSClient        stsiface.STSAPI
	asgName          string
	launchConfigName string
}

func (c *ClientSet) DiscoverClients() error {
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
	log.Infof("Credentials: %v", arn)

	c.ASClient = autoscaling.New(sess)
	c.EKSClient = eks.New(sess)
	c.Route53Client = route53.New(sess)
	c.IAMClient = iam.New(sess)
	c.STSClient = sts.New(sess)

	return nil
}

func (c *ClientSet) AnASGNamed(name string) error {
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
	log.Infof("Auto Scaling group: %v", arn)

	c.launchConfigName = aws.StringValue(out.AutoScalingGroups[0].LaunchConfigurationName)
	c.asgName = name

	return nil
}

func (c *ClientSet) ScaleCurrentASG(desiredMin, desiredMax int64) error {

	if c.ASClient == nil {
		return errors.Errorf("Unable to scale currrent ASG: The AS client was not found, use the method GetAWSCredsAndClients")
	}

	_, err := c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(c.asgName),
		MinSize:              aws.Int64(desiredMin),
		MaxSize:              aws.Int64(desiredMax),
	})
	if err != nil {
		return errors.Errorf("Failed scaling ASG %v: %v", c.asgName, err)
	}

	return nil
}

func (c *ClientSet) UpdateFieldOfCurrentASG(field, value string) error {
	var (
		err        error
		valueInt64 int64
	)

	if c.ASClient == nil {
		return errors.Errorf("Unable to update current ASG: The AS client was not found, use the method GetAWSCredsAndClients")
	}

	if field == "LaunchConfigurationName" {
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName:    aws.String(c.asgName),
			LaunchConfigurationName: aws.String(value),
		})

		if err != nil {
			return errors.Errorf("Failed updating field %v to %v of ASG %v: %v", field, value, c.asgName, err)
		}
		return nil
	}

	valueInt64, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return errors.Errorf("Failed to convert %v to int64 while trying to update field %v of ASG %v", value, field, c.asgName)
	}

	switch field {
	case "MinSize":
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(c.asgName),
			MinSize:              aws.Int64(valueInt64),
		})
	case "DesiredCapacity":
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(c.asgName),
			DesiredCapacity:      aws.Int64(valueInt64),
		})
	case "MaxSize":
		_, err = c.ASClient.UpdateAutoScalingGroup(&autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(c.asgName),
			MaxSize:              aws.Int64(valueInt64),
		})
	default:
		return errors.Errorf("Field %v is not supported, use LaunchConfigurationName, MinSize, DesiredCapacity or MaxSize", field)
	}

	if err != nil {
		return errors.Errorf("Failed updating field %v to %v of ASG %v: %v", field, value, c.asgName, err)
	}
	return nil
}

func (c *ClientSet) IamRoleTrust(action, entityName, roleName string) error {
	accountId := getAccountNumber(c.STSClient)
	clusterName, err := getClusterName()
	if err != nil {
		return err
	}

	// Add efs-csi-role-<clustername> as trusted entity
	var trustedEntityArn = fmt.Sprintf("arn:aws:iam::%s:role/%s-%s",
		accountId, entityName, clusterName)

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

	role, err := kIam.GetIamRole(roleName, c.IAMClient)
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

	_, err = kIam.UpdateIAMAssumeRole(roleName, policyJSON, c.IAMClient)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClientSet) ClusterSharedIamOperation(operation string) error {
	var (
		accountId = getAccountNumber(c.STSClient)
		iamFmt    = "arn:aws:iam::%s:%s/%s"
	)
	clusterName, err := getClusterName()
	if err != nil {
		return err
	}
	roleName := fmt.Sprintf("shared.%s", clusterName)

	policyDocT := `{"Version":"2012-10-17","Statement":[{"Effect": "Allow", "Action": "sts:AssumeRole", "Resource": "%s"}]}`
	clusterSharedrole := fmt.Sprintf(iamFmt, accountId, "role", roleName)
	policyDocument := []byte(fmt.Sprintf(policyDocT, clusterSharedrole))

	rootIAM := fmt.Sprintf("arn:aws:iam::%s:%s", accountId, "root")
	assumeRoleDoc := `{"Version":"2012-10-17","Statement":[{"Effect": "Allow", "Action": "sts:AssumeRole", "Principal": {"AWS": "%s"}}]}`
	roleDocument := []byte(fmt.Sprintf(assumeRoleDoc, rootIAM))

	clusterSharedPolicy := fmt.Sprintf(iamFmt, accountId, "policy", roleName)
	switch operation {
	case "add":
		role, err := kIam.PutIAMRole(roleName, "shared cluster role", roleDocument, c.IAMClient)
		if err != nil {
			return errors.Wrap(err, "failed to create shared cluster role")
		}
		log.Infof("BDD >> created shared iam role: %s", aws.StringValue(role.Arn))

		policy, err := kIam.PutManagedPolicy(roleName, clusterSharedPolicy, "shared cluster policy", policyDocument, c.IAMClient)
		if err != nil {
			return errors.Wrap(err, "failed to create shared cluster managed policy")
		}
		log.Infof("BDD >> created shared iam policy: %s", aws.StringValue(policy.Arn))
	case "remove":
		err := kIam.DeleteManagedPolicy(clusterSharedPolicy, c.IAMClient)
		if err != nil {
			return errors.Wrap(err, "failed to delete shared cluster role")
		}

		err = kIam.DeleteIAMRole(roleName, c.IAMClient)
		if err != nil {
			return errors.Wrap(err, "failed to delete shared cluster managed policy")
		}
	}
	return nil
}

func (c *ClientSet) DnsNameShouldOrNotInHostedZoneID(dnsName, shouldOrNot, hostedZoneID string) error {
	switch shouldOrNot {
	case "should":
		return c.dnsNameInHostedZoneID(dnsName, hostedZoneID)

	case "should not":
		if err := c.dnsNameInHostedZoneID(dnsName, hostedZoneID); err == nil {
			return errors.Errorf("unexpected DNS %s exists in hostedZoneID %s", hostedZoneID, dnsName)
		}
		log.Infof("records for hostedZoneID %s with dnsName %s doesn't exists", hostedZoneID, dnsName)
		return nil
	default:
		return fmt.Errorf("invalid option '%s'. expected 'should' or 'should not'", shouldOrNot)
	}
}
