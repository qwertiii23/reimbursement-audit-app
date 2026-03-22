<template>
  <div class="condition-group">
    <draggable
      v-model="localConditions"
      item-key="id"
      handle=".drag-handle"
      @end="handleDragEnd"
    >
      <template #item="{ element: item, index }">
        <div class="condition-item">
          <div class="condition-logic" v-if="index > 0">
            <el-select
              v-model="item.logicOp"
              size="small"
              style="width: 60px"
            >
              <el-option label="AND" value="and" />
              <el-option label="OR" value="or" />
            </el-select>
          </div>

          <div class="condition-content-wrapper">
            <div v-if="item.type === 'group'" class="condition-group-box">
              <div class="group-header">
                <span class="group-label">条件组</span>
                <el-button
                  type="danger"
                  link
                  :icon="Delete"
                  size="small"
                  @click="removeItem(index)"
                >
                  删除组
                </el-button>
              </div>
              <div class="group-conditions">
                <ConditionGroup
                  :conditions="item.conditions"
                  :features="features"
                  @update="(data) => updateItem(index, data)"
                  @add-condition="() => addConditionToGroup(index)"
                  @add-group="() => addGroupToGroup(index)"
                  @remove="(childIndex) => removeItemFromGroup(index, childIndex)"
                />
              </div>
              <div class="group-actions">
                <el-button type="primary" link size="small" @click="addConditionToGroup(index)">
                  添加条件
                </el-button>
                <el-button type="primary" link size="small" @click="addGroupToGroup(index)">
                  添加条件组
                </el-button>
              </div>
            </div>

            <div v-else class="condition-box">
              <div class="condition-row">
                <div class="drag-handle">
                  <el-icon><Rank /></el-icon>
                </div>
                <span class="condition-label">条件{{ index + 1 }}:</span>
                <el-select
                  v-model="item.featureId"
                  placeholder="选择特征"
                  style="width: 150px"
                  filterable
                  @change="handleFeatureChange(index)"
                >
                  <el-option
                    v-for="feature in features"
                    :key="feature.id"
                    :label="feature.name"
                    :value="feature.id"
                  />
                </el-select>
                <el-select
                  v-model="item.operator"
                  placeholder="操作符"
                  style="width: 100px"
                  :disabled="!item.featureId"
                >
                  <el-option label="=" value="eq" />
                  <el-option label="!=" value="ne" />
                  <el-option label=">" value="gt" />
                  <el-option label=">=" value="gte" />
                  <el-option label="<" value="lt" />
                  <el-option label="<=" value="lte" />
                  <el-option label="包含" value="contains" />
                  <el-option label="不包含" value="not_contains" />
                </el-select>
                <el-select
                  v-if="getFeatureType(item.featureId) === 'string'"
                  v-model="item.value"
                  placeholder="选择特征值"
                  style="width: 150px"
                  filterable
                  allow-create
                  :disabled="!item.featureId"
                >
                  <el-option
                    v-for="value in getFeatureValues(item.featureId)"
                    :key="value.value"
                    :label="value.label"
                    :value="value.value"
                  />
                </el-select>
                <el-input
                  v-else
                  v-model="item.value"
                  placeholder="请输入数值"
                  style="width: 120px"
                  :disabled="!item.featureId"
                />
                <el-button
                  type="primary"
                  link
                  size="small"
                  @click="addGroupAfter(index)"
                >
                  添加组
                </el-button>
                <el-button
                  type="danger"
                  link
                  :icon="Delete"
                  size="small"
                  @click="removeItem(index)"
                >
                  删除
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </draggable>
  </div>
</template>

<script setup>
import { Delete, Plus, Rank } from '@element-plus/icons-vue'
import { computed } from 'vue'
import draggable from 'vuedraggable'

const props = defineProps({
  conditions: {
    type: Array,
    default: () => []
  },
  features: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update', 'add-condition', 'add-group', 'remove'])

const localConditions = computed({
  get: () => props.conditions,
  set: (value) => {
    emit('update', value)
  }
})

const handleDragEnd = () => {
  emit('update', localConditions.value)
}

const handleFeatureChange = (index) => {
  const newConditions = [...props.conditions]
  const item = newConditions[index]
  const feature = props.features.find(f => f.id === item.featureId)
  if (feature) {
    if (feature.type === 'string' && feature.values && feature.values.length > 0) {
      item.value = feature.values[0].value
    } else if (feature.type === 'number') {
      item.value = ''
    } else {
      item.value = ''
    }
  } else {
    item.value = ''
  }
  emit('update', newConditions)
}

const updateItem = (index, data) => {
  const newConditions = [...props.conditions]
  newConditions[index] = { ...newConditions[index], ...data }
  emit('update', newConditions)
}

const removeItem = (index) => {
  const newConditions = [...props.conditions]
  newConditions.splice(index, 1)
  emit('update', newConditions)
}

const addGroupAfter = (index) => {
  const newConditions = [...props.conditions]
  newConditions.splice(index + 1, 0, {
    id: '',
    type: 'group',
    logicOp: 'and',
    conditions: []
  })
  emit('update', newConditions)
}

const addConditionToGroup = (groupIndex) => {
  const newConditions = [...props.conditions]
  const group = newConditions[groupIndex]
  group.conditions.push({
    id: '',
    type: 'condition',
    featureId: '',
    operator: 'eq',
    value: '',
    logicOp: group.conditions.length > 0 ? 'and' : 'and'
  })
  emit('update', newConditions)
}

const addGroupToGroup = (groupIndex) => {
  const newConditions = [...props.conditions]
  const group = newConditions[groupIndex]
  group.conditions.push({
    id: '',
    type: 'group',
    logicOp: group.conditions.length > 0 ? 'and' : 'and',
    conditions: []
  })
  emit('update', newConditions)
}

const removeItemFromGroup = (groupIndex, childIndex) => {
  const newConditions = [...props.conditions]
  const group = newConditions[groupIndex]
  group.conditions.splice(childIndex, 1)
  emit('update', newConditions)
}

const getFeatureValues = (featureId) => {
  const feature = props.features.find(f => f.id === featureId)
  return feature && feature.values ? feature.values : []
}

const getFeatureType = (featureId) => {
  const feature = props.features.find(f => f.id === featureId)
  return feature ? feature.type : 'string'
}
</script>

<style scoped>
.condition-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.condition-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.condition-logic {
  display: flex;
  align-items: center;
  min-width: 70px;
  padding-top: 5px;
}

.condition-content-wrapper {
  flex: 1;
}

.condition-group-box {
  padding: 10px;
  background-color: #f8f9fa;
  border: 1px solid #409eff;
  border-radius: 4px;
  margin-bottom: 5px;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 5px;
  border-bottom: 1px solid #e0e0e0;
}

.group-label {
  font-weight: 600;
  color: #409eff;
  font-size: 13px;
}

.group-conditions {
  margin-bottom: 8px;
  padding-left: 10px;
}

.group-actions {
  display: flex;
  gap: 10px;
  padding-top: 5px;
  border-top: 1px solid #e0e0e0;
}

.condition-box {
  padding: 8px;
  background-color: #f5f7fa;
  border-radius: 4px;
  border-left: 2px solid #409eff;
}

.condition-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.drag-handle {
  cursor: move;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: #909399;
  transition: color 0.3s;
}

.drag-handle:hover {
  color: #409eff;
}

.condition-label {
  font-weight: 500;
  color: #606266;
  font-size: 13px;
  min-width: 50px;
}
</style>
