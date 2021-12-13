package workerapi

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
)

// Fofa Fofa任务
func Fofa(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.FofaConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	fofa := onlineapi.NewFofa(config)
	fofa.Do()
	if config.IsIPLocation {
		doLocation(&fofa.IpResult)
	}
	//指纹识别：
	if len(fofa.IpResult.IPResult) > 0 {
		portscanConfig := portscan.Config{
			IsHttpx:          config.IsHttpx,
			IsWhatWeb:        config.IsWhatWeb,
			IsWappalyzer:     config.IsWappalyzer,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		DoIPFingerPrint(portscanConfig, &fofa.IpResult)
		if fofa.Config.IsScreenshot {
			DoScreenshotAndSave(&fofa.IpResult, nil)
		}
	}
	if len(fofa.DomainResult.DomainResult) > 0 {
		domainscanConfig := domainscan.Config{
			IsHttpx:          config.IsHttpx,
			IsWhatWeb:        config.IsWhatWeb,
			IsWappalyzer:     config.IsWappalyzer,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		DoDomainFingerPrint(domainscanConfig, &fofa.DomainResult)
		if fofa.Config.IsScreenshot {
			DoScreenshotAndSave(nil, &fofa.DomainResult)
		}
	}
	// 保存结果
	x := comm.NewXClient()
	args := comm.ScanResultArgs{
		TaskID:       taskId,
		IPConfig:     &portscan.Config{OrgId: config.OrgId},
		DomainConfig: &domainscan.Config{OrgId: config.OrgId},
		IPResult:     fofa.IpResult.IPResult,
		DomainResult: fofa.DomainResult.DomainResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &args, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}
