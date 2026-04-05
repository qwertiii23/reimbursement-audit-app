#!/bin/bash

# 完整的报销审核流程测试脚本

BASE_URL="http://localhost:8080"
TRACE_ID="test-$(date +%s)"

echo "=========================================="
echo "报销审核流程测试"
echo "=========================================="
echo "Trace ID: $TRACE_ID"
echo ""

# 步骤1：创建报销单
echo "步骤1：创建报销单"
echo "----------------------------------------"
curl -X POST "$BASE_URL/api/v1/reimbursement/upload" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d '{
    "user_id": "test-user-001",
    "user_name": "张三",
    "department": "技术部",
    "category": "交通费",
    "reason": "上海出差交通费",
    "description": "从北京到上海出差，乘坐高铁往返",
    "total_amount": 1500.00,
    "apply_date": "2026-03-01",
    "expense_date": "2026-03-01",
    "start_date": "2026-03-05",
    "end_date": "2026-03-07",
    "destination": "上海",
    "city": "上海",
    "province": "上海",
    "travel_reason": "参加技术会议",
    "transportation": "高铁",
    "applicant_level": "普通员工"
  }' | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤2：上传发票
echo "步骤2：上传发票"
echo "----------------------------------------"
curl -X POST "$BASE_URL/api/v1/invoices/upload" \
  -H "Content-Type: multipart/form-data" \
  -H "X-Trace-ID: $TRACE_ID" \
  -F "file=@test_invoice.jpg" \
  -F "reimbursement_id=test-reimbursement-001" \
  -F "invoice_code=1234567890" \
  -F "invoice_number=00012345" \
  -F "invoice_date=2026-03-01" \
  -F "amount=1500.00" \
  -F "merchant_type=交通" \
  -F "merchant_name=中国铁路上海局" \
  -F "invoice_type=增值税专用发票" | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤3：触发OCR解析
echo "步骤3：触发OCR解析"
echo "----------------------------------------"
curl -X POST "$BASE_URL/api/v1/invoices/ocr" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d '{
    "invoice_id": "test-invoice-001"
  }' | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤4：开始智能审核
echo "步骤4：开始智能审核"
echo "----------------------------------------"
curl -X POST "$BASE_URL/api/v1/audit" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d '{
    "reimbursement_id": "test-reimbursement-001",
    "audit_type": "smart"
  }' | jq '.'

echo ""
echo "等待5秒，等待规则引擎执行..."
sleep 5

# 步骤5：查询审核状态
echo "步骤5：查询审核状态"
echo "----------------------------------------"
curl -X GET "$BASE_URL/api/v1/audit/status/test-audit-001" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤6：查询审核结果
echo "步骤6：查询审核结果"
echo "----------------------------------------"
curl -X GET "$BASE_URL/api/v1/audit/result/test-audit-001" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤7：流转到人工审核
echo "步骤7：流转到人工审核"
echo "----------------------------------------"
curl -X PUT "$BASE_URL/api/v1/reimbursement/test-reimbursement-001" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d '{
    "status": "manual_review",
    "audit_comment": "智能审核通过，需要人工复核"
  }' | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤8：人工审核通过
echo "步骤8：人工审核通过"
echo "----------------------------------------"
curl -X PUT "$BASE_URL/api/v1/reimbursement/test-reimbursement-001" \
  -H "Content-Type: application/json" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d '{
    "status": "approved",
    "audit_comment": "人工审核通过",
    "approved_by": "admin"
  }' | jq '.'

echo ""
echo "等待2秒..."
sleep 2

# 步骤9：查询最终状态
echo "步骤9：查询最终状态"
echo "----------------------------------------"
curl -X GET "$BASE_URL/api/v1/reimbursement/test-reimbursement-001" \
  -H "X-Trace-ID: $TRACE_ID" | jq '.'

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
