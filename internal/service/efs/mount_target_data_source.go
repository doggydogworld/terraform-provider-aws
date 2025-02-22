// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_efs_mount_target")
func DataSourceMountTarget() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceMountTargetRead,

		Schema: map[string]*schema.Schema{
			"access_point_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_target_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mount_target_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMountTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn(ctx)

	input := &efs.DescribeMountTargetsInput{}

	if v, ok := d.GetOk("access_point_id"); ok {
		input.AccessPointId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_system_id"); ok {
		input.FileSystemId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("mount_target_id"); ok {
		input.MountTargetId = aws.String(v.(string))
	}

	mt, err := findMountTarget(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target: %s", err)
	}

	d.SetId(aws.StringValue(mt.MountTargetId))
	fsID := aws.StringValue(mt.FileSystemId)
	fsARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  "file-system/" + fsID,
		Service:   "elasticfilesystem",
	}.String()
	d.Set("availability_zone_id", mt.AvailabilityZoneId)
	d.Set("availability_zone_name", mt.AvailabilityZoneName)
	d.Set("dns_name", meta.(*conns.AWSClient).RegionalHostname(fsID+".efs"))
	d.Set("file_system_arn", fsARN)
	d.Set("file_system_id", fsID)
	d.Set("ip_address", mt.IpAddress)
	d.Set("mount_target_dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.%s.efs", aws.StringValue(mt.AvailabilityZoneName), aws.StringValue(mt.FileSystemId))))
	d.Set("mount_target_id", mt.MountTargetId)
	d.Set("network_interface_id", mt.NetworkInterfaceId)
	d.Set("owner_id", mt.OwnerId)
	d.Set("subnet_id", mt.SubnetId)

	output, err := conn.DescribeMountTargetSecurityGroupsWithContext(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s) security groups: %s", d.Id(), err)
	}

	d.Set("security_groups", aws.StringValueSlice(output.SecurityGroups))

	return diags
}
