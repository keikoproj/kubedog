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
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	clusterNameEnvironmentVariable = "CLUSTER_NAME"
)

func (c *ClientSet) GetEksVpc() (string, error) {
	clusterName, err := getClusterName()
	if err != nil {
		return "", err
	}
	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}
	result, err := c.EKSClient.DescribeCluster(input)
	if err != nil {
		return "", err
	}
	return aws.StringValue(result.Cluster.ResourcesVpcConfig.VpcId), nil
}

func getAccountNumber(svc stsiface.STSAPI) string {
	// Region is defaulted to "us-west-2"
	input := &sts.GetCallerIdentityInput{}
	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		log.Infof("Failed to get caller identity: %s", err.Error())
		return ""
	}
	return *result.Account
}

func (c *ClientSet) getDNSRecord(dnsName string, hostedZoneID string) (string, error) {
	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(hostedZoneID),
		MaxItems:        aws.String("1"),
		StartRecordName: aws.String(dnsName),
	}
	resp, err := c.Route53Client.ListResourceRecordSets(params)
	if err != nil {
		return "", err
	}
	if len(resp.ResourceRecordSets) == 0 {
		return "", fmt.Errorf("no record set exists for hostedZoneID %v with dnsName %v", hostedZoneID, dnsName)
	}
	recordSet := resp.ResourceRecordSets[0]

	if recordSet.AliasTarget != nil {
		aliasRecordValue := aws.StringValue(recordSet.AliasTarget.DNSName)
		if aliasRecordValue == "" {
			return "", errors.New(fmt.Sprintf("no record set exists for hostedZoneID %v with dnsName %v", hostedZoneID, dnsName))
		}
		return aliasRecordValue, nil
	} else {
		if len(recordSet.ResourceRecords) == 0 {
			return "", errors.New(fmt.Sprintf("no record set exists for hostedZoneID %v with dnsName %v", hostedZoneID, dnsName))
		}

		recordValue := aws.StringValue(recordSet.ResourceRecords[0].Value)
		if recordValue == "" {
			return "", fmt.Errorf("no record set exists for hostedZoneID %v with dnsName %v", hostedZoneID, dnsName)
		}
		return recordValue, nil
	}
}

func (c *ClientSet) dnsNameInHostedZoneID(dnsName, hostedZoneID string) error {
	recordValue, err := c.getDNSRecord(dnsName, hostedZoneID)
	if err != nil {
		if recordValue != "" {
			log.Infof("records for hostedZoneID %s with dnsName %s exists", hostedZoneID, dnsName)
			return nil
		} else {
			return errors.Errorf("records for hostedZoneID %s with dnsName %s doesn't exists", hostedZoneID, dnsName)
		}
	}
	log.Infof("records for hostedZoneID %s with dnsName %s exists", hostedZoneID, dnsName)
	return nil
}

func getClusterName() (string, error) {
	return getEnv(clusterNameEnvironmentVariable)
}

func getEnv(envName string) (string, error) {
	if envValue, ok := os.LookupEnv(envName); ok {
		return envValue, nil
	}
	return "", fmt.Errorf("could not get environment variable '%s'", envName)
}
