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
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	// TODO: support multiple ASG
	ASClient         autoscalingiface.AutoScalingAPI
	AsgName          string
	LaunchConfigName string
}

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
	}

	ASGs := out.AutoScalingGroups
	switch len(ASGs) {
	case 1:
		arn := aws.StringValue(ASGs[0].AutoScalingGroupARN)
		log.Infof("[KUBEDOG] Auto Scaling group: %v", arn)
		c.LaunchConfigName = aws.StringValue(ASGs[0].LaunchConfigurationName)
		c.AsgName = name
		return nil
	case 0:
		return errors.Errorf("No ASG found by the name: '%s'", name)
	default:
		// Not likely to happen. Here in case something inherently wrong with AWS/API
		return errors.Errorf("DescribeAutoScalingGroups returned more than 1 ASG with the name '%s': %v", name, ASGs)
	}
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
