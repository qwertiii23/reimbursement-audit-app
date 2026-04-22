#!/bin/bash

# 简化的报销审核流程测试脚本

BASE_URL="http://localhost:8080"
REIMBURSEMENT_ID="26c046d6-da56-4802-87b9-b3f6c6256f58"
TRACE_ID="test-$(date +%s)"

echo "=========================================="
echo "报销审核流程测试（简化版）"
echo "=========================================="
echo "Trace ID: $TRACE_ID"
echo "报销单ID: $REIMBURSEMENT_ID"
echo ""

# 步骤1：查询报销单详情
echo "步骤1：查询报销单详情"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/api/v1/reimbursement/$REIMBURSEMENT_ID" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "等待1秒..."
sleep 1

# 步骤2：开始智能审核
echo "步骤2：开始智能审核"
echo "----------------------------------------"
curl -s -X POST "$BASE_URL/api/v1/audit" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d "{
    \"reimbursement_id\": \"$REIMBURSEMENT_ID\",
    \"audit_type\": \"smart\"
  }" | jq '.'

echo ""
echo "等待3秒，等待规则引擎执行..."
sleep 3

# 步骤3：查询审核状态
echo "步骤3：查询审核状态"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/api/v1/audit/status/$REIMBURSEMENT_ID" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "等待1秒..."
sleep 1

# 步骤4：查询审核结果
echo "步骤4：查询审核结果"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/api/v1/audit/result/$REIMBURSEMENT_ID" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "等待1秒..."
sleep 1

# 步骤5：流转到人工审核
echo "步骤5：流转到人工审核"
echo "----------------------------------------"
curl -s -X PUT "$BASE_URL/api/v1/reimbursement/$REIMBURSEMENT_ID" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d "{
    \"status\": \"manual_review\",
    \"audit_comment\": \"智能审核完成，需要人工复核\"
  }" | jq '.'

echo ""
echo "等待1秒..."
sleep 1

# 步骤6：人工审核通过
echo "步骤6：人工审核通过"
echo "----------------------------------------"
curl -s -X PUT "$BASE_URL/api/v1/reimbursement/$REIMBURSEMENT_ID" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d "{
    \"status\": \"approved\",
    \"audit_comment\": \"人工审核通过，所有规则校验正常\",
    \"approved_by\": \"admin\"
  }" | jq '.'

echo ""
echo "等待1秒..."
sleep 1

# 步骤7：查询最终状态
echo "步骤7：查询最终状态"
echo "----------------------------------------"
curl -s -X GET "$BASE_URL/api/v1/reimbursement/$REIMBURSEMENT_ID" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
