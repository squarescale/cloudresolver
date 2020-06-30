package cloudresolver

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-05-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/davecgh/go-spew/spew"
	"os"
	"strings"
)

type AzureResolver struct{}

type IpInfos struct {
	PrivateIpv4  string
	PrivateIpv6  string
	PrivateName  string
	PublicIpv4   string
	PublicIpv6   string
	PublicName   string
}

func init() {
	register("azure", new(AzureResolver))
}

func getAuthorizer() (autorest.Authorizer, error) {
	//authorizer, err := auth.NewAuthorizerFromCLI()
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

func getVMClient() (*compute.VirtualMachinesClient, error) {
	vmClient := compute.NewVirtualMachinesClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := getAuthorizer()
	if err != nil {
		return nil, err
	}
	vmClient.Authorizer = authorizer
	return &vmClient, nil
}

// GetAllVMs gets all VMs info
func GetAllVMs(ctx context.Context) (*compute.VirtualMachineListResultIterator, error) {
	vmClient, err := getVMClient()
	if err != nil {
		return nil, err
	}
	res, ret1 := vmClient.ListAllComplete(ctx, "false")
	return &res, ret1
}


func getVMSSClient() (*compute.VirtualMachineScaleSetsClient, error) {
	vmSSClient := compute.NewVirtualMachineScaleSetsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := getAuthorizer()
	if err != nil {
		return nil, err
	}
	vmSSClient.Authorizer = authorizer
	return &vmSSClient, nil
}

// GetAllVMScaleSets gets all VM Scale Sets info
func GetAllVMSSs(ctx context.Context) (*compute.VirtualMachineScaleSetListWithLinkResultIterator, error) {
	vmSSClient, err := getVMSSClient()
	if err != nil {
		return nil, err
	}
	res, ret1 := vmSSClient.ListAllComplete(ctx)
	return &res, ret1
}

func getVMSSVMsClient() (*compute.VirtualMachineScaleSetVMsClient, error) {
	vmSSVMsClient := compute.NewVirtualMachineScaleSetVMsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := getAuthorizer()
	if err != nil {
		return nil, err
	}
	vmSSVMsClient.Authorizer = authorizer
	return &vmSSVMsClient, nil
}

// GetAllVMScaleSets gets all VM Scale Sets instances info
func getAllVMSSsVMsInstances(ctx context.Context, id *string) (*compute.VirtualMachineScaleSetVMListResultIterator, error) {
	vmSSVMsClient, err := getVMSSVMsClient()
	if err != nil {
		return nil, err
	}
	vmssID := strings.Split(*id, "/")
	res, ret1 := vmSSVMsClient.ListComplete(ctx, vmssID[4], vmssID[8], "", "", "")
	return &res, ret1
}

func getNetITFClient() (*network.InterfacesClient, error) {
	netITFClient := network.NewInterfacesClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := getAuthorizer()
	if err != nil {
		return nil, err
	}
	netITFClient.Authorizer = authorizer
	return &netITFClient, nil
}

func getPubIPCfgClient() (*network.PublicIPAddressesClient, error) {
	pubIPCfgClient := network.NewPublicIPAddressesClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := getAuthorizer()
	if err != nil {
		return nil, err
	}
	pubIPCfgClient.Authorizer = authorizer
	return &pubIPCfgClient, nil
}

func getNetIPCfgClient() (*network.InterfaceIPConfigurationsClient, error) {
	netIPCfgClient := network.NewInterfaceIPConfigurationsClient(os.Getenv("AZURE_SUBSCRIPTION_ID"))
	authorizer, err := getAuthorizer()
	if err != nil {
		return nil, err
	}
	netIPCfgClient.Authorizer = authorizer
	return &netIPCfgClient, nil
}

// GetNetworkInterfaceInfos gets infos on specific network interface
func GetNetworkInterfaceInfos(ctx context.Context, reference compute.NetworkInterfaceReference) (*IpInfos, error) {
	ipi := IpInfos{}
	netITFClient, err := getNetITFClient()
	if err != nil {
		return nil, err
	}
	/*netIPCfgClient, err := getNetIPCfgClient()
	if err != nil {
		return nil, err
	}*/
	pubIPCfgClient, err := getPubIPCfgClient()
	if err != nil {
		return nil, err
	}
	netITFID := strings.Split(*reference.ID, "/")
	res, err := netITFClient.Get(ctx, netITFID[4], netITFID[8], "")
	if err != nil {
		return nil, err
	}
	for _, ipcfg := range *res.IPConfigurations {
		//fmt.Printf("x: %s\n", spew.Sdump(ipcfg))
		pubITFID := strings.Split(*ipcfg.PublicIPAddress.ID, "/")
		if *ipcfg.Primary && ipcfg.PrivateIPAddressVersion == network.IPv4 {
			ipi.PrivateIpv4 = *ipcfg.PrivateIPAddress
		}
		if *ipcfg.Primary && ipcfg.PrivateIPAddressVersion == network.IPv6 {
			ipi.PrivateIpv6 = *ipcfg.PrivateIPAddress
		}
		pubipcfg, err := pubIPCfgClient.Get(ctx, pubITFID[4], pubITFID[8], "")
		if err != nil {
			return nil, err
		}
		if pubipcfg.PublicIPAddressVersion == network.IPv4 {
			ipi.PublicIpv4 = *pubipcfg.IPAddress
		}
		if pubipcfg.PublicIPAddressVersion == network.IPv6 {
			ipi.PublicIpv6 = *pubipcfg.IPAddress
		}
		if pubipcfg.DNSSettings != nil {
			ipi.PublicName = *pubipcfg.DNSSettings.Fqdn
		}
		//fmt.Printf("x: %s\n", spew.Sdump(pubipcfg.PublicIPAddressPropertiesFormat))
	}
	return &ipi, nil
}

// GetNetworkInterfaceInfos gets infos on specific network interface
func GetVMSSInstanceNetworkInterfaceInfos(ctx context.Context, reference compute.NetworkInterfaceReference) (*IpInfos, error) {
	ipi := IpInfos{}
	netITFClient, err := getNetITFClient()
	if err != nil {
		return nil, err
	}
	netITFID := strings.Split(*reference.ID, "/")
	res, err := netITFClient.GetVirtualMachineScaleSetNetworkInterface(ctx, netITFID[4], netITFID[8], netITFID[10], netITFID[12], "")
	if err != nil {
		return nil, err
	}
	for _, ipcfg := range *res.IPConfigurations {
		//fmt.Printf("x: %s\n", spew.Sdump(ipcfg))
		if *ipcfg.Primary && ipcfg.PrivateIPAddressVersion == network.IPv4 {
			ipi.PrivateIpv4 = *ipcfg.PrivateIPAddress
		}
		if *ipcfg.Primary && ipcfg.PrivateIPAddressVersion == network.IPv6 {
			ipi.PrivateIpv6 = *ipcfg.PrivateIPAddress
		}
	}
	return &ipi, nil
}
func (r AzureResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	//ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	//defer cancel()  // releases resources if slowOperation completes before timeout elapses
	// Virtual Machines
	allVMs, err := GetAllVMs(context.Background())
	if err != nil {
		fmt.Printf("All VMs error: %s\n", spew.Sdump(err))
		return []Host{}, err
	}
	hosts := []Host{}
	for vmListItem := allVMs; vmListItem.NotDone(); err = vmListItem.NextWithContext(context.Background()) {
		vm := vmListItem.Value()
		//fmt.Printf("VM: %s", spew.Sdump(vm))
		tags := make(map[string]string)
		iname := ""
		for k, v := range vm.Tags {
			if k == "Name" {
				iname = *v
			}
			tags[k] = *v
		}
		if len(name) > 0 {
			if name != "*" && !strings.HasPrefix(iname, name) {
				continue
			}
		}
		h := Host{
			InstanceName: iname,
			MachineType:  string(vm.VirtualMachineProperties.HardwareProfile.VMSize),
			Provider:     "azure",
			Region:       *vm.Location,
			//Zone:         *inst.Placement.AvailabilityZone,
			Id:           *vm.ID,
			Tags:         tags,
		}
		for _, ni := range *vm.NetworkProfile.NetworkInterfaces {
			niInfos, err := GetNetworkInterfaceInfos(context.Background(), ni)
			if err != nil {
				fmt.Printf("VM Network interface error: %s\n", spew.Sdump(err))
				return []Host{}, err
			}
			h.PrivateIpv4 = niInfos.PrivateIpv4
			h.Private     = niInfos.PrivateIpv4
			h.PublicIpv4  = niInfos.PublicIpv4
			h.Public      = niInfos.PublicIpv4
			h.PublicName  = niInfos.PublicName
		}
		hosts = append(hosts, h)
	}
	// Virtual Machines Scale Sets
	allVMSSs, err := GetAllVMSSs(context.Background())
	if err != nil {
		fmt.Printf("VM scale sets error: %s\n", spew.Sdump(err))
		return []Host{}, err
	}
	for vmssListItem := allVMSSs; vmssListItem.NotDone(); err = vmssListItem.NextWithContext(context.Background()) {
		vmss := vmssListItem.Value()
		//fmt.Printf("VMSS: ==============================\n%s------------------------------\n", spew.Sdump(vmss))
		tags := make(map[string]string)
		iname := ""
		for k, v := range vmss.Tags {
			if k == "Name" {
				iname = *v
			}
			tags[k] = *v
		}
		if vmss.VirtualMachineProfile != nil && vmss.VirtualMachineProfile.OsProfile != nil && vmss.VirtualMachineProfile.OsProfile.ComputerNamePrefix != nil {
			iname = *vmss.VirtualMachineProfile.OsProfile.ComputerNamePrefix
		}
		// Virtual Machines Scale Sets Instances
		allVMSSInstances, err := getAllVMSSsVMsInstances(context.Background(), vmss.ID)
		if err != nil {
			fmt.Printf("IP error: %s\n", spew.Sdump(err))
			return []Host{}, err
		}
		for instListItem := allVMSSInstances; instListItem.NotDone(); err = instListItem.NextWithContext(context.Background()) {
			vmssInstance := instListItem.Value()
			//fmt.Printf("VMSSInstance: ==============================\n%s------------------------------\n", spew.Sdump(vmssInstance))
			for k, v := range vmssInstance.Tags {
				if k == "Name" {
					iname = *v + "_" + *vmssInstance.InstanceID
				}
				if tags[k] != *v {
					fmt.Printf("Replacing VMSS tag %s from %s to %s", k, tags[k], *v)
					tags[k] = *v
				}
			}
			if len(name) > 0 {
				if name != "*" && !strings.HasPrefix(iname, name) && !strings.HasPrefix(*vmss.Name, name){
					continue
				}
			}
			h := Host{
				InstanceName: iname,
				MachineType:  *vmssInstance.Sku.Name,
				Provider:     "azure",
				Region:       *vmssInstance.Location,
				//Zone:         *inst.Placement.AvailabilityZone,
				Id:           *vmssInstance.ID,
				Tags:         tags,
			}
			for _, ni := range *vmssInstance.NetworkProfile.NetworkInterfaces {
				//fmt.Printf("Got %s", spew.Sdump(ni))
				niInfos, err := GetVMSSInstanceNetworkInterfaceInfos(context.Background(), ni)
				if err != nil {
					fmt.Printf("VM Scale Set Instance Network interface error: %s\n", spew.Sdump(err))
					return []Host{}, err
				}
				h.PrivateIpv4 = niInfos.PrivateIpv4
				h.Private     = niInfos.PrivateIpv4
				/*h.PublicIpv4  = niInfos.PublicIpv4
				h.Public      = niInfos.PublicIpv4
				h.PublicName  = niInfos.PublicName*/
			}
			hosts = append(hosts, h)
		}
	}
	return hosts, nil
}
