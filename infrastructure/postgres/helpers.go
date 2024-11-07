package infrastructure

import (
	"encoding/json"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	domain_audit_log "ifttt/handler/domain/audit_log"
	"ifttt/handler/domain/resolvable"

	"github.com/jackc/pgtype"
)

func (pgRule *rules) toDomain() (*api.Rule, error) {
	domainRule := api.Rule{
		ID:          pgRule.ID,
		Name:        pgRule.Name,
		Description: pgRule.Description,
	}

	if err := json.Unmarshal(pgRule.Pre.Bytes, &domainRule.Pre); err != nil {
		return nil,
			fmt.Errorf("method *PostgresRulesRepository.ToDomain: error in unmarshalling pre: %s", err)
	}

	if err := json.Unmarshal(pgRule.Switch.Bytes, &domainRule.Switch); err != nil {
		return nil,
			fmt.Errorf("method *PostgresRulesRepository.ToDomain: error in unmarshalling switch: %s", err)
	}

	return &domainRule, nil
}

func (t *trigger_flows) toDomain() (*api.TriggerFlow, error) {
	domanTFlow := api.TriggerFlow{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Class:       api.Class{Name: t.Class.Name},
		StartState:  t.StartState,
		Rules:       map[uint]*api.Rule{},
		BranchFlows: map[uint]*api.BranchFlow{},
	}
	for _, r := range t.Rules {
		dRule, err := r.toDomain()
		if err != nil {
			return nil,
				fmt.Errorf("method *PostgresTriggerFlowsRepository.ToDomain: could not convert to domain rule: %s", err)
		}
		domanTFlow.Rules[r.ID] = dRule
	}

	if err := json.Unmarshal(t.BranchFlows.Bytes, &domanTFlow.BranchFlows); err != nil {
		return nil,
			fmt.Errorf("method *PostgresRulesRepository.ToDomain: error in unmarshalling branchFlows: %s", err)
	}

	return &domanTFlow, nil
}

func (a *apis) toDomain() (*api.Api, error) {
	domainApi := api.Api{
		ID:          a.ID,
		Name:        a.Name,
		Path:        a.Path,
		Method:      a.Method,
		Description: a.Description,
		Request:     map[string]any{},
		PreConfig:   map[string]resolvable.Resolvable{},
		PreWare:     &[]api.TriggerFlow{},
		MainWare:    &[]api.TriggerCondition{},
		PostWare:    &[]api.TriggerFlow{},
	}

	if err := json.Unmarshal(a.Request.Bytes, &domainApi.Request); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(a.PreConfig.Bytes, &domainApi.PreConfig); err != nil {
		return nil, err
	}

	var tConditions []api_trigger_flow_json
	if err := json.Unmarshal(a.TriggerFlows.Bytes, &tConditions); err != nil {
		return nil, err
	}

	for _, tFlow := range a.PreWare {
		domainTFlow, err := tFlow.toDomain()
		if err != nil {
			return nil, err
		}
		*domainApi.PreWare = append(*domainApi.PreWare, *domainTFlow)
	}

	for _, tFlow := range a.PostWare {
		domainTFlow, err := tFlow.toDomain()
		if err != nil {
			return nil, err
		}
		*domainApi.PostWare = append(*domainApi.PostWare, *domainTFlow)
	}

	triggerFlowMap := make(map[uint]trigger_flows)
	for _, tFlow := range a.MainWare {
		triggerFlowMap[tFlow.ID] = tFlow
	}

	for _, tc := range tConditions {
		tcModel, ok := triggerFlowMap[tc.Trigger]
		if !ok {
			return nil,
				fmt.Errorf("method *PostgresAPIRepository.ToDomain: trigger flow not found from conditions")
		}
		domainTFlow, err := tcModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("method *PostgresAPIRepository.ToDomain: %s", err)
		}
		*domainApi.MainWare = append(*domainApi.MainWare,
			api.TriggerCondition{If: tc.If, Trigger: *domainTFlow})
	}

	return &domainApi, nil
}

func (a *api_audit_log) fromDomain(dLog *domain_audit_log.APIAuditLog) error {
	a.ApiID = dLog.ApiID
	a.ApiName = dLog.ApiName
	a.ApiPath = dLog.ApiPath
	a.Start = dLog.Start
	a.End = dLog.End
	a.ExecTime = dLog.ExecTime
	a.InternalExecTime = dLog.InternalExecTime
	a.ExternalExecTime = dLog.ExternalExecTime
	a.ResponseSent = dLog.ResponseSent
	a.RequestToken = dLog.RequestToken

	if execOrderMarshalled, err := json.Marshal(common.UnSyncMap(dLog.ExecutionOrder)); err != nil {
		return fmt.Errorf("method *PostgresAPIRepository.FromDomain: could not marshal execution order: %s", err)
	} else {
		a.ExecutionOrder = pgtype.JSONB{Bytes: execOrderMarshalled, Status: pgtype.Present}
	}

	if execLogsMarshalled, err := json.Marshal(dLog.ExecutionLogs); err != nil {
		return fmt.Errorf("method *PostgresAPIRepository.FromDomain: could not marshal execution logs: %s", err)
	} else {
		a.ExecutionLogs = pgtype.JSONB{Bytes: execLogsMarshalled, Status: pgtype.Present}
	}

	reqDataMap := make(map[string]any)
	reqDataMap["reqBody"] = dLog.RequestData.ReqBody
	reqDataMap["preConfig"] = common.UnSyncMap(dLog.RequestData.PreConfig)
	reqDataMap["store"] = common.UnSyncMap(dLog.RequestData.Store)
	reqDataMap["response"] = common.UnSyncMap(dLog.RequestData.Response)
	reqDataMap["queryRes"] = dLog.RequestData.QueryRes
	reqDataMap["apiRes"] = dLog.RequestData.ApiRes
	if reqDataMarshalled, err := json.Marshal(&reqDataMap); err != nil {
		return fmt.Errorf("method *PostgresAPIRepository.FromDomain: could not marshal request data: %s", err)
	} else {
		a.RequestData = pgtype.JSONB{Bytes: reqDataMarshalled, Status: pgtype.Present}
	}

	if finalResMarshalled, err := json.Marshal(&dLog.FinalResponse); err != nil {
		return err
	} else {
		a.FinalResponse = pgtype.JSONB{Bytes: finalResMarshalled, Status: pgtype.Present}
	}

	return nil
}

func (c *crons) toDomain() (*api.Cron, error) {
	dCron := api.Cron{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Name,
		Cron:         c.Cron,
		TriggerFlows: &[]api.TriggerCondition{},
	}

	if err := json.Unmarshal(c.PreConfig.Bytes, &dCron.PreConfig); err != nil {
		return nil,
			fmt.Errorf("method *PostgresAPIRepository.ToDomain: could not cast pgApi: %s", err)
	}

	var tConditions []api_trigger_flow_json
	if err := json.Unmarshal(c.TriggerFlows.Bytes, &tConditions); err != nil {
		return nil,
			fmt.Errorf("method *PostgresAPIRepository.ToDomain: could not cast pgApi: %s", err)
	}

	triggerFlowMap := make(map[uint]trigger_flows)
	for _, tFlow := range c.TriggerFlowRef {
		triggerFlowMap[tFlow.ID] = tFlow
	}

	for _, tc := range tConditions {
		tcModel, ok := triggerFlowMap[tc.Trigger]
		if !ok {
			return nil,
				fmt.Errorf("method *PostgresAPIRepository.ToDomain: trigger flow not found from conditions")
		}
		domainTFlow, err := tcModel.toDomain()
		if err != nil {
			return nil, fmt.Errorf("method *PostgresAPIRepository.ToDomain: %s", err)
		}
		*dCron.TriggerFlows = append(*dCron.TriggerFlows, api.TriggerCondition{If: tc.If, Trigger: *domainTFlow})
	}

	return &dCron, nil
}

func (a *cron_audit_log) fromDomain(dLog *domain_audit_log.CronAuditLog) error {
	a.CronName = dLog.Name
	a.CronID = dLog.CronID
	a.Start = dLog.Start
	a.End = dLog.End
	a.ExecTime = dLog.ExecTime
	a.InternalExecTime = dLog.InternalExecTime
	a.ExternalExecTime = dLog.ExternalExecTime
	a.ResponseSent = dLog.ResponseSent
	a.RequestToken = dLog.RequestToken

	if execOrderMarshalled, err := json.Marshal(common.UnSyncMap(dLog.ExecutionOrder)); err != nil {
		return fmt.Errorf("method *PostgresAPIRepository.FromDomain: could not marshal execution order: %s", err)
	} else {
		a.ExecutionOrder = pgtype.JSONB{Bytes: execOrderMarshalled, Status: pgtype.Present}
	}

	if execLogsMarshalled, err := json.Marshal(dLog.ExecutionLogs); err != nil {
		return fmt.Errorf("method *PostgresAPIRepository.FromDomain: could not marshal execution logs: %s", err)
	} else {
		a.ExecutionLogs = pgtype.JSONB{Bytes: execLogsMarshalled, Status: pgtype.Present}
	}

	reqDataMap := make(map[string]any)
	reqDataMap["reqBody"] = dLog.RequestData.ReqBody
	reqDataMap["preConfig"] = common.UnSyncMap(dLog.RequestData.PreConfig)
	reqDataMap["store"] = common.UnSyncMap(dLog.RequestData.Store)
	reqDataMap["response"] = common.UnSyncMap(dLog.RequestData.Response)
	reqDataMap["queryRes"] = dLog.RequestData.QueryRes
	reqDataMap["apiRes"] = dLog.RequestData.ApiRes
	if reqDataMarshalled, err := json.Marshal(&reqDataMap); err != nil {
		return err
	} else {
		a.RequestData = pgtype.JSONB{Bytes: reqDataMarshalled, Status: pgtype.Present}
	}

	if finalResMarshalled, err := json.Marshal(&dLog.FinalResponse); err != nil {
		return err
	} else {
		a.FinalResponse = pgtype.JSONB{Bytes: finalResMarshalled, Status: pgtype.Present}
	}

	return nil
}
