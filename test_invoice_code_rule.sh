#!/bin/bash

# 测试发票代码格式校验规则

echo "测试1: 发票代码为10位纯数字（应该通过）"
curl -X POST http://localhost:8080/api/v1/engine/test \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_code": "1234567890",
    "invoice_no": "00123456",
    "amount": 100.00
  }' | jq '.'

echo ""
echo "测试2: 发票代码为12位纯数字（应该通过）"
curl -X POST http://localhost:8080/api/v1/engine/test \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_code": "123456789012",
    "invoice_no": "00123456",
    "amount": 100.00
  }' | jq '.'

echo ""
echo "测试3: 发票代码为9位纯数字（应该不通过）"
curl -X POST http://localhost:8080/api/v1/engine/test \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_code": "123456789",
    "invoice_no": "00123456",
    "amount": 100.00
  }' | jq '.'

echo ""
echo "测试4: 发票代码为13位纯数字（应该不通过）"
curl -X POST http://localhost:8080/api/v1/engine/test \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_code": "1234567890123",
    "invoice_no": "00123456",
    "amount": 100.00
  }' | jq '.'

echo ""
echo "测试5: 发票代码包含字母（应该不通过）"
curl -X POST http://localhost:8080/api/v1/engine/test \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_code": "1234567890A",
    "invoice_no": "00123456",
    "amount": 100.00
  }' | jq '.'
