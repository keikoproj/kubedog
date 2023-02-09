package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *Client) GetDNSRecord(dnsName string, hostedZoneID string) (string, error) {
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

func (c *Client) DnsNameInHostedZoneID(dnsName, hostedZoneID string) error {
	recordValue, err := c.GetDNSRecord(dnsName, hostedZoneID)
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
