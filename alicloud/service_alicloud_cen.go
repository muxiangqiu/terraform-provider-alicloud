package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cbn"
)

const DefaultCenTimeout = 60
const DefaultCenTimeoutLong = 180

func (client *AliyunClient) DescribeCenInstance(cenId string) (c cbn.Cen, err error) {
	request := cbn.CreateDescribeCensRequest()

	values := []string{cenId}
	filters := []cbn.DescribeCensFilter{cbn.DescribeCensFilter{
		Key:   "CenId",
		Value: &values,
	}}

	request.Filter = &filters

	invoker := NewInvoker()
	err = invoker.Run(func() error {
		resp, err := client.cenconn.DescribeCens(request)
		if err != nil {
			if IsExceptedError(err, ParameterCenInstanceIdNotExist) {
				return GetNotFoundErrorFromString(GetNotFoundMessage("CEN Instance", cenId))
			}
			return err
		}
		if resp == nil || len(resp.Cens.Cen) <= 0 || resp.Cens.Cen[0].CenId != cenId {
			return GetNotFoundErrorFromString(GetNotFoundMessage("CEN Instance", cenId))
		}
		c = resp.Cens.Cen[0]
		return nil
	})

	return
}

func (client *AliyunClient) WaitForCenInstance(cenId string, status Status, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		cen, err := client.DescribeCenInstance(cenId)
		if err != nil {
			return err
		}
		if cen.Status == string(status) {
			break
		}
		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(GetTimeoutMessage("CEN", string(status)))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) DescribeCenAttachedChildInstanceById(instanceId, cenId string) (c cbn.ChildInstance, err error) {
	request := cbn.CreateDescribeCenAttachedChildInstancesRequest()
	request.CenId = cenId

	for pageNum := 1; ; pageNum++ {
		request.PageNumber = requests.NewInteger(pageNum)
		response, err := client.cenconn.DescribeCenAttachedChildInstances(request)
		if err != nil {
			return c, err
		}

		instanceList := response.ChildInstances.ChildInstance
		for instanceNum := 0; instanceNum <= len(instanceList)-1; instanceNum++ {
			if instanceList[instanceNum].ChildInstanceId == instanceId {
				return instanceList[instanceNum], nil
			}
		}

		if pageNum*response.PageSize >= response.TotalCount {
			return c, GetNotFoundErrorFromString(GetNotFoundMessage("CEN Child Instance", instanceId))
		}
	}
}

func (client *AliyunClient) WaitForCenChildInstanceAttached(instanceId string, cenId string, status Status, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		instance, err := client.DescribeCenAttachedChildInstanceById(instanceId, cenId)
		if err != nil {
			return err
		}
		if instance.Status == string(status) {
			break
		}
		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(GetTimeoutMessage("CEN Child Instance Attachment", string(status)))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) WaitForCenChildInstanceDetached(instanceId string, cenId string, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		_, err := client.DescribeCenAttachedChildInstanceById(instanceId, cenId)
		if err != nil {
			if NotFoundError(err) {
				break
			} else {
				return err
			}
		}

		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(fmt.Sprintf("Waitting for %s detach timeout.", instanceId))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) DescribeCenBandwidthPackage(cenBwpId string) (c cbn.CenBandwidthPackage, err error) {
	request := cbn.CreateDescribeCenBandwidthPackagesRequest()

	values := []string{cenBwpId}
	filters := []cbn.DescribeCenBandwidthPackagesFilter{cbn.DescribeCenBandwidthPackagesFilter{
		Key:   "CenBandwidthPackageId",
		Value: &values,
	}}
	request.Filter = &filters

	invoker := NewInvoker()
	err = invoker.Run(func() error {
		resp, err := client.cenconn.DescribeCenBandwidthPackages(request)
		if err != nil {
			if IsExceptedError(err, ParameterCenInstanceIdNotExist) {
				return GetNotFoundErrorFromString(GetNotFoundMessage("CEN Bandwidth Package", cenBwpId))
			}
			return err
		}
		if resp == nil || len(resp.CenBandwidthPackages.CenBandwidthPackage) <= 0 || resp.CenBandwidthPackages.CenBandwidthPackage[0].CenBandwidthPackageId != cenBwpId {
			return GetNotFoundErrorFromString(GetNotFoundMessage("CEN Bandwidth Package", cenBwpId))
		}
		c = resp.CenBandwidthPackages.CenBandwidthPackage[0]
		return nil
	})

	return
}

func (client *AliyunClient) WaitForCenBandwidthPackage(cenBwpId string, status Status, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		cenBwp, err := client.DescribeCenBandwidthPackage(cenBwpId)
		if err != nil && !NotFoundError(err) {
			return err
		}
		if cenBwp.Status == string(status) {
			break
		}
		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(GetTimeoutMessage("CEN Bandwidth Package", string(status)))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) WaitForCenBandwidthPackageUpdate(cenBwpId string, bandwidth int, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		cenBwp, err := client.DescribeCenBandwidthPackage(cenBwpId)
		if err != nil {
			return err
		}
		if cenBwp.Bandwidth == bandwidth {
			break
		}

		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(fmt.Sprintf("Waitting for CEN bandwidth package update is timeout"))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) WaitForCenBandwidthPackageAttachment(cenBwpId string, status Status, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		cenBwp, err := client.DescribeCenBandwidthPackage(cenBwpId)
		if err != nil {
			return err
		}
		if cenBwp.Status == string(status) {
			break
		}
		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(GetTimeoutMessage("CEN Bandwidth Package Attachment", string(status)))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) DescribeCenBandwidthPackageById(cenBwpId string) (c cbn.CenBandwidthPackage, err error) {
	resp, err := client.DescribeCenBandwidthPackage(cenBwpId)
	if err != nil {
		return c, err
	}

	if len(resp.CenIds.CenId) != 1 || resp.Status != string(InUse) {
		return c, GetNotFoundErrorFromString(GetNotFoundMessage("CEN bandwidth package attachment", cenBwpId))
	}

	return resp, nil
}

func (client *AliyunClient) SetCenInterRegionBandwidthLimit(cenId, localRegionId, oppositeRegionId string, bandwidthLimit int) (err error) {
	request := cbn.CreateSetCenInterRegionBandwidthLimitRequest()
	request.CenId = cenId
	request.LocalRegionId = localRegionId
	request.OppositeRegionId = oppositeRegionId
	request.BandwidthLimit = requests.NewInteger(bandwidthLimit)

	_, err = client.cenconn.SetCenInterRegionBandwidthLimit(request)

	return err
}

func (client *AliyunClient) DescribeCenBandwidthLimit(cenId, localRegionId, oppositeRegionId string) (c cbn.CenInterRegionBandwidthLimit, err error) {
	request := cbn.CreateDescribeCenInterRegionBandwidthLimitsRequest()
	request.CenId = cenId

	for pageNum := 1; ; pageNum++ {
		request.PageNumber = requests.NewInteger(pageNum)
		request.PageSize = requests.NewInteger(PageSizeLarge)
		resp, err := client.cenconn.DescribeCenInterRegionBandwidthLimits(request)
		if err != nil {
			return c, err
		}

		cenBandwidthLimitList := resp.CenInterRegionBandwidthLimits.CenInterRegionBandwidthLimit
		for limitNum := 0; limitNum <= len(cenBandwidthLimitList)-1; limitNum++ {
			ifMatch := cenBandwidthLimitList[limitNum].LocalRegionId == localRegionId && cenBandwidthLimitList[limitNum].OppositeRegionId == oppositeRegionId
			if !ifMatch {
				ifMatch = cenBandwidthLimitList[limitNum].LocalRegionId == oppositeRegionId && cenBandwidthLimitList[limitNum].OppositeRegionId == localRegionId
			}
			if ifMatch {
				return cenBandwidthLimitList[limitNum], nil
			}
		}

		if pageNum*resp.PageSize >= resp.TotalCount {
			return c, GetNotFoundErrorFromString(fmt.Sprintf("The specified CEN bandwith limit CEN Id %s localRegionId %s oppositeRegionId %s is not found", cenId, localRegionId, oppositeRegionId))
		}
	}
}

func (client *AliyunClient) WaitForCenInterRegionBandwidthLimitActive(cenId string, localRegionId string, oppositeRegionId string, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		cenBandwidthLimit, err := client.DescribeCenBandwidthLimit(cenId, localRegionId, oppositeRegionId)
		if err != nil {
			return err
		}

		if cenBandwidthLimit.Status == string(Active) {
			break
		}

		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(fmt.Sprintf("Waitting for bandwidth limit CenId %s localRegionId %s oppositeRegionId %s timeout.", cenId, localRegionId, oppositeRegionId))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func (client *AliyunClient) WaitForCenInterRegionBandwidthLimitDestroy(cenId string, localRegionId string, oppositeRegionId string, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	for {
		_, err := client.DescribeCenBandwidthLimit(cenId, localRegionId, oppositeRegionId)
		if err != nil {
			if NotFoundError(err) {
				break
			}
			return err
		}

		timeout = timeout - DefaultIntervalShort
		if timeout <= 0 {
			return GetTimeErrorFromString(fmt.Sprintf("Waitting for bandwidth limit CenId %s localRegionId %s oppositeRegionId %s timeout.", cenId, localRegionId, oppositeRegionId))
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}

	return nil
}

func getCenIdAndAnotherId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource id")
	}

	return parts[0], parts[1], nil
}

func getCenAndRegionIds(id string) (retString []string, err error) {
	parts := strings.Split(id, ":")

	if len(parts) != 3 {
		return retString, fmt.Errorf("invalid resource id")
	}

	return parts, nil
}

func getCenInstanceType(id string) (c string, e error) {
	if strings.HasPrefix(id, "vpc") {
		return "VPC", nil
	} else if strings.HasPrefix(id, "vbr") {
		return "VBR", nil
	} else {
		return c, fmt.Errorf("CEN child instance ID invalid. Now, it only supports VPC or VBR instance.")
	}
}
