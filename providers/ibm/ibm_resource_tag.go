package ibm

import (
	"fmt"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/IBM/go-sdk-core/core"
	"github.com/IBM/vpc-go-sdk/vpcv1"
)

// ResourceTagGenerator ...
type ResourceTagGenerator struct {
	IBMService
}

func (g ResourceTagGenerator) createTaggingResources(resID, resName string) terraformutils.Resource {
	resource := terraformutils.NewResource(
		resID,
		normalizeResourceName(resName, true),
		"ibm_resource_tag",
		"ibm",
		map[string]string{},
		[]string{},
		map[string]interface{}{})

	return resource
}

// InitResources ...
func (g *ResourceTagGenerator) InitResources() error {
	region := g.Args["region"].(string)
	apiKey := os.Getenv("IC_API_KEY")
	if apiKey == "" {
		log.Fatal("No API key set")
	}

	isURL := GetVPCEndPoint(region)
	iamURL := GetAuthEndPoint()
	vpcoptions := &vpcv1.VpcV1Options{
		URL: isURL,
		Authenticator: &core.IamAuthenticator{
			ApiKey: apiKey,
			URL:    iamURL,
		},
	}
	vpcclient, err := vpcv1.NewVpcV1(vpcoptions)
	if err != nil {
		return err
	}
	start := ""
	var allrecs []vpcv1.FloatingIP
	for {
		options := &vpcv1.ListFloatingIpsOptions{}
		if start != "" {
			options.Start = &start
		}
		if rg := g.Args["resource_group"].(string); rg != "" {
			rg, err = GetResourceGroupID(apiKey, rg, region)
			if err != nil {
				return fmt.Errorf("Error Fetching Resource Group Id %s", err)
			}
			options.ResourceGroupID = &rg
		}
		fips, response, err := vpcclient.ListFloatingIps(options)
		if err != nil {
			return fmt.Errorf("Error Fetching Floating IPs %s\n%s", err, response)
		}
		start = GetNext(fips.Next)
		allrecs = append(allrecs, fips.FloatingIps...)
		if start == "" {
			break
		}
	}

	for _, fip := range allrecs {
		if fip.Target != nil {
			g.Resources = append(g.Resources, g.createTaggingResources(*fip.ID, *fip.Name))
		}
	}

	return nil
}
