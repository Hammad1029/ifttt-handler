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
		StartRules:  []uint{},
		AllRules:    map[uint]*api.Rule{},
		BranchFlows: map[uint]*[]api.BranchFlow{},
	}
	for _, r := range t.StartRules {
		domanTFlow.StartRules = append(domanTFlow.StartRules, r.ID)
	}
	for _, r := range t.AllRules {
		dRule, err := r.toDomain()
		if err != nil {
			return nil,
				fmt.Errorf("method *PostgresTriggerFlowsRepository.ToDomain: could not convert to domain rule")
		}
		domanTFlow.AllRules[r.ID] = dRule
	}

	if err := json.Unmarshal(t.BranchFlow.Bytes, &domanTFlow.BranchFlows); err != nil {
		return nil,
			fmt.Errorf("method *PostgresRulesRepository.ToDomain: error in unmarshalling branchFlows: %s", err)
	}

	return &domanTFlow, nil
}

func (a *apis) toDomain() (*api.Api, error) {
	domainApi := api.Api{
		ID:           a.ID,
		Name:         a.Name,
		Path:         a.Path,
		Method:       a.Method,
		Description:  a.Description,
		Request:      map[string]any{},
		PreConfig:    map[string]resolvable.Resolvable{},
		TriggerFlows: &[]api.TriggerCondition{},
	}

	if err := json.Unmarshal(a.Request.Bytes, &domainApi.Request); err != nil {
		return nil,
			fmt.Errorf("method *PostgresAPIRepository.ToDomain: could not cast pgApi: %s", err)
	}

	if err := json.Unmarshal(a.PreConfig.Bytes, &domainApi.PreConfig); err != nil {
		return nil,
			fmt.Errorf("method *PostgresAPIRepository.ToDomain: could not cast pgApi: %s", err)
	}

	var tConditions []api_trigger_flow_json
	if err := json.Unmarshal(a.TriggerFlows.Bytes, &tConditions); err != nil {
		return nil,
			fmt.Errorf("method *PostgresAPIRepository.ToDomain: could not cast pgApi: %s", err)
	}

	triggerFlowMap := make(map[uint]trigger_flows)
	for _, tFlow := range a.TriggerFlowRef {
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
		*domainApi.TriggerFlows = append(*domainApi.TriggerFlows,
			api.TriggerCondition{If: tc.If, Trigger: *domainTFlow})
	}

	return &domainApi, nil
}

func (a *audit_log) fromDomain(dLog *domain_audit_log.AuditLog) error {
	a.ApiID = dLog.ApiID
	a.ApiName = dLog.ApiName
	a.ApiPath = dLog.ApiPath
	a.Start = dLog.Start
	a.End = dLog.End
	a.TimeTaken = dLog.TimeTaken

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

	return nil
}
