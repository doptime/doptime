package dynamicdev

// roleProductManager = "1.负责定义产品的愿景和方向。2.理解市场需求，制定产品路线图。3.与客户和用户沟通，收集反馈，并根据反馈调整产品功能。"
type MetaProductManager struct {
	ProductVision                 string `json:"product_vision"`
	MarcketDemand                 string `json:"marcket_demand"`
	RefactorDemandsOfUserFeedback string `json:"refactor_demands_of_user_feedback"`
}

var ArchitectGlobal string = "使用golang doptime为主要框架来进行开发。apiINfo, dataInfo, OtherFunctionNames"

type ApiSchema struct {
	ParamIn  string `json:"param_in"`
	ParamOut string `json:"param_out"`
}
type DataSchema struct {
	RedisType string `json:"redis_type"`
	KeyType   string `json:"key_type"`
	ValueType string `json:"value_type"`
}

// roleArchitect = "1.设计系统的整体结构，确保架构满足业务需求并具有可扩展性。2.选择合适的技术栈和开发工具。3.解决技术难题，指导开发团队在技术实施方面。"
type MetaArchitect struct {
	ApiSchemas           []*ApiSchema  `json:"api_schemas,omitempty"`
	DataSchemas          []*DataSchema `json:"data_schema,omitempty"`
	OtherFunctionSchemas []string      `json:"function_schema,omitempty"`
	RefactorRequirements []string      `json:"refactor_requirements"`
}

// roleEngineer = "1.编写代码实现产品功能。2.参与代码审查，确保代码质量。3.与其他团队成员合作，如测试工程师和架构师，以确保软件的整体质量和一致性。"

type MetaEngineerIn struct {
	CurrentCode string `json:"current_code"`
	VerboseLog  string `json:"verbose_log"`
	ExampleCode string `json:"example_code"`
}

// Step-level Beam Search MCTS.use MCTS to make revision
// https://openi.cn/142542.html
// https://m.thepaper.cn/newsDetail_forward_27309947
type MetaEngineerOut struct {
	EvaluationOfCurrentCode string `json:"evaluation_of_current_code"`
	ExtensionOfCurrentCode  string `json:"extension_of_current_code"`
	PruningOfCurrentCode    string `json:"pruning_of_current_code"`
}

type MetaStatus struct {
	VerboseLog string `json:"verbose_log"`
}
