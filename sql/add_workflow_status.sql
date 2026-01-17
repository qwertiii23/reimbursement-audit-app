-- 添加workflow_status字段到audit_results表
ALTER TABLE `audit_results` ADD COLUMN `workflow_status` VARCHAR(32) DEFAULT '已提交' COMMENT '工作流状态' AFTER `status`;
ALTER TABLE `audit_results` ADD INDEX `idx_workflow_status` (`workflow_status`);
