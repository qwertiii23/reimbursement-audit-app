-- 报销制度文档表
-- 用于存储报销制度文档的分片内容和向量嵌入

CREATE TABLE IF NOT EXISTS reimbursement_documents (
    id VARCHAR(36) PRIMARY KEY COMMENT '文档ID',
    file_name VARCHAR(255) NOT NULL COMMENT '文件名',
    file_type VARCHAR(20) NOT NULL COMMENT '文件类型（pdf/word/txt）',
    chunk_id VARCHAR(36) NOT NULL COMMENT '分片ID',
    chunk_index INT NOT NULL COMMENT '分片序号',
    chunk_content TEXT NOT NULL COMMENT '分片内容',
    embedding TEXT COMMENT '向量（JSON+BASE64）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_chunk_id (chunk_id),
    INDEX idx_file_name (file_name),
    INDEX idx_file_type (file_type),
    FULLTEXT idx_chunk_content (chunk_content) COMMENT '全文索引，用于文本匹配检索'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='报销制度文档表';

-- 插入demo数据（基于reimbursement_policy_demo.txt）
-- 这里先插入几个示例分片，实际应用中应该通过文档处理器自动分片

-- 分片1：第一章 总则
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-001', 'reimbursement_policy_demo.txt', 'txt', 'chunk-001', 1, 
'第一章 总则

第一条 为规范公司财务报销行为，加强费用管理，提高资金使用效益，根据国家有关财经法规和公司实际情况，制定本制度。

第二条 本制度适用于公司全体员工因公发生的各项费用报销。

第三条 报销原则：真实性、合法性、必要性、合理性、节约性。',
NOW());

-- 分片2：第二章 差旅费报销 - 交通费标准
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-002', 'reimbursement_policy_demo.txt', 'txt', 'chunk-002', 2, 
'第二章 差旅费报销

第五条 交通费标准
1. 飞机票：公司高管（总经理、副总经理等）可根据工作需要选择商务舱，其他员工一律乘坐经济舱。特殊情况需乘坐商务舱或头等舱的，必须提前报请总经理批准。
2. 火车票：员工出差应优先选择高铁或动车二等座，特殊情况可选择一等座。普通列车硬座为标准，硬卧为特殊情况。
3. 轮船票：三等舱为标准，二等舱为特殊情况。
4. 长途汽车：按实际票价报销。
5. 市内交通：出差期间市内交通费按每人每天50元标准包干使用，凭票报销，超支不补。',
NOW());

-- 分片3：第二章 差旅费报销 - 住宿费标准
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-003', 'reimbursement_policy_demo.txt', 'txt', 'chunk-003', 3, 
'第二章 差旅费报销

第六条 住宿费标准
1. 一线城市（北京、上海、广州、深圳）：普通员工每人每天500元，部门经理每人每天800元，高管每人每天1200元。
2. 二线城市（省会城市、计划单列市）：普通员工每人每天400元，部门经理每人每天600元，高管每人每天1000元。
3. 三线城市及其他地区：普通员工每人每天300元，部门经理每人每天500元，高管每人每天800元。
4. 两人以上同性别员工出差，原则上应合住标准间，住宿费按一人标准报销。
5. 超标住宿需提前报请部门负责人批准，超出部分由个人承担。',
NOW());

-- 分片4：第二章 差旅费报销 - 伙食补助费标准
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-004', 'reimbursement_policy_demo.txt', 'txt', 'chunk-004', 4, 
'第二章 差旅费报销

第七条 伙食补助费标准
1. 一线城市：每人每天150元。
2. 二线城市：每人每天120元。
3. 三线城市及其他地区：每人每天100元。
4. 伙食补助费按出差天数计算，不足半天的按半天计算，超过半天的按全天计算。
5. 出差期间如由接待单位提供用餐的，相应天数不享受伙食补助。',
NOW());

-- 分片5：第三章 业务招待费报销
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-005', 'reimbursement_policy_demo.txt', 'txt', 'chunk-005', 5, 
'第三章 业务招待费报销

第十条 业务招待费是指为开展业务活动需要而发生的宴请、礼品、娱乐等费用。

第十一条 招待对象分类
1. A类客人：公司重要客户、合作伙伴、政府官员等。招待标准为餐费350-400元/人/次，住宿五星级以上。
2. B类客人：一般客户、供应商等。招待标准为餐费200-300元/人/次，住宿四星级。
3. C类客人：其他业务相关人员。招待标准为餐费100-150元/人/次，住宿三星级。

第十二条 招待费标准
1. 餐费：根据招待对象分类，按上述标准执行。
2. 住宿费：根据招待对象分类，按上述标准执行。
3. 礼品费：一般不超过500元/人，特殊情况需总经理批准。
4. 娱乐费：一般不超过300元/人，特殊情况需总经理批准。',
NOW());

-- 分片6：第三章 业务招待费报销 - 审批和税务
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-006', 'reimbursement_policy_demo.txt', 'txt', 'chunk-006', 6, 
'第三章 业务招待费报销

第十三条 招待费审批
1. 招待前必须填写《招待申请表》，注明招待对象、人数、预算、事由等，经部门负责人批准。
2. 单次招待费超过2000元的，需报请分管领导批准。
3. 单次招待费超过5000元的，需报请总经理批准。

第十五条 招待费税务规定
根据国家税务总局规定，企业发生的与生产经营活动有关的业务招待费支出，按照发生额的60%扣除，但最高不得超过当年销售（营业）收入的5‰。',
NOW());

-- 分片7：第四章 办公费报销
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-007', 'reimbursement_policy_demo.txt', 'txt', 'chunk-007', 7, 
'第四章 办公费报销

第十六条 办公费是指为维持公司正常办公所发生的各项费用，包括办公用品、办公设备、通讯费、网络费、快递费等。

第十七条 办公用品报销
1. 办公用品由行政部统一采购，员工不得私自采购。
2. 特殊情况需自行采购的，需提前报请行政部批准。
3. 办公用品报销需提供发票和采购清单。
4. 单次办公用品采购超过1000元的，需行政部负责人批准。
5. 单次办公用品采购超过5000元的，需分管领导批准。

第十八条 办公设备报销
1. 办公设备由行政部统一采购，员工不得私自采购。
2. 办公设备包括电脑、打印机、复印机、办公家具等。
3. 办公设备采购需填写《设备采购申请表》，经部门负责人、行政部负责人、分管领导批准。
4. 办公设备采购超过10000元的，需报请总经理批准。',
NOW());

-- 分片8：第五章 培训费报销
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-008', 'reimbursement_policy_demo.txt', 'txt', 'chunk-008', 8, 
'第五章 培训费报销

第二十一条 培训费是指为提高员工业务能力和综合素质而发生的培训费用。

第二十二条 培训费标准
1. 内部培训：讲师费每人每天500-1000元，场地费每人每天50-100元，材料费每人每天20-50元。
2. 外部培训：培训费根据培训机构收费标准执行，一般不超过5000元/人/次。
3. 境外培训：培训费根据实际情况确定，需报请总经理批准。

第二十三条 培训费审批
1. 培训前必须填写《培训申请表》，注明培训内容、时间、地点、人数、预算等。
2. 内部培训需部门负责人批准。
3. 外部培训需部门负责人和分管领导批准。
4. 境外培训需总经理批准。',
NOW());

-- 分片9：第六章 会议费报销
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-009', 'reimbursement_policy_demo.txt', 'txt', 'chunk-009', 9, 
'第六章 会议费报销

第二十五条 会议费是指为召开会议而发生的各项费用，包括场地费、餐饮费、住宿费、交通费、材料费等。

第二十六条 会议费标准
1. 场地费：根据场地档次和规模确定，一般不超过5000元/天。
2. 餐饮费：每人每天100-150元。
3. 住宿费：按差旅费住宿费标准执行。
4. 交通费：按差旅费交通费标准执行。
5. 材料费：每人每天20-50元。

第二十七条 会议费审批
1. 会议前必须填写《会议申请表》，注明会议主题、时间、地点、人数、预算等。
2. 部门级会议需部门负责人批准。
3. 公司级会议需分管领导批准。
4. 大型会议需总经理批准。',
NOW());

-- 分片10：第十章 报销流程
INSERT INTO reimbursement_documents (id, file_name, file_type, chunk_id, chunk_index, chunk_content, created_at) VALUES
('doc-001-chunk-010', 'reimbursement_policy_demo.txt', 'txt', 'chunk-010', 10, 
'第十章 报销流程

第四十二条 报销申请
1. 员工填写《费用报销单》，注明费用类型、金额、事由、时间等。
2. 附上相关凭证（发票、申请表、清单等）。
3. 提交部门负责人审核。

第四十四条 审批权限
1. 1000元以下：部门负责人审批。
2. 1000-5000元：部门负责人和财务负责人审批。
3. 5000-10000元：部门负责人、财务负责人和分管领导审批。
4. 10000元以上：部门负责人、财务负责人、分管领导和总经理审批。

第四十五条 报销时限
1. 员工应在费用发生后5个工作日内提交报销申请。
2. 超过报销时限的，需说明原因并经财务负责人批准。
3. 超过30天未报销的，原则上不予报销。',
NOW());