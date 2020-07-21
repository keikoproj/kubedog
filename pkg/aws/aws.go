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

func (c *Client) AnASGNamed(name string) error {

	sess, err := GetAWSCredentials()
	if err != nil {
		return errors.Errorf("Failed getting AWS credentials: %v", err)
	}

	ASC := autoscaling.New(sess)

	// Simple call to check valid client
	out, err := ASC.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
	})
	if err != nil {
		return errors.Errorf("Failed describing the ASG %v: %v", name, err)
	}

	arn := *out.AutoScalingGroups[0].AutoScalingGroupARN
	log.Infof("BDD >> Auto Scaling group: %v", arn)

	c.LaunchConfigName = *out.AutoScalingGroups[0].LaunchConfigurationName
	c.ASClient = ASC
	c.AsgName = name

	return nil
}

func (c *Client) ScaleCurrentASG(desiredMin, desiredMax int64) error {

	if c.ASClient == nil {
		return errors.Errorf("Unable to scale currrent ASG, no ASG client found in kubedog.Test.AwsContext")
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
Fields/Parameters supported: LaunchConfigurationName, MinSize,DesiredCapacity and MaxSize
*/
func (c *Client) UpdateFieldOfCurrentASG(field, value string) error {
	var (
		err        error
		valueInt64 int64
	)

	if c.ASClient == nil {
		return errors.Errorf("The ASG client was not found, use the method AnASGNamed")
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

func GetAWSCredentials() (*session.Session, error) {
	var (
		sess     *session.Session
		identity *sts.GetCallerIdentityOutput
		err      error
	)

	if sess, err = session.NewSession(); err != nil {
		return nil, err
	}

	svc := sts.New(sess)

	if identity, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{}); err != nil {
		return nil, err
	}

	arn := aws.StringValue(identity.Arn)
	log.Infof("BDD >> Credentials: %v", arn)

	return sess, nil
}
