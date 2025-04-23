<template>
  <div class="app-container">
    <el-container>
      <el-main class="left-panel">

        <el-row :gutter="20" class="top-cards">
          <el-col :span="12">
            <el-card class="patient-search-card">
              <template #header>
                <div class="patient-search-header">
                  <el-icon><UserSearch /></el-icon>
                  <span>病人 ID 检索</span>
                </div>
              </template>
              <div class="patient-search-input">
                <el-input
                    v-model="currentPatientId"
                    placeholder="输入病人 ID（如 patient_123）"
                    class="input-with-button"
                >
                  <template #append>
                    <el-button type="primary" @click="fetchPatientImages">检索</el-button>
                  </template>
                </el-input>
              </div>
            </el-card>
          </el-col>

          <el-col :span="12">
            <el-card class="timer-card">
              <template #header>
                <div class="timer-header">
                  <el-icon><Clock /></el-icon>
                  <span>计时器</span>
                </div>
              </template>
              <div class="timer-content">
                <span class="time-display">{{ time }} S</span>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-card class="selected-images-card">
          <template #header>
            <div class="selected-images-header">选中的图片</div>
          </template>
          <div class="selected-images-scroll">
            <div class="selected-images">
              <img
                  v-for="(imgUrl, index) in selectedImages"
                  :key="index"
                  :src="imgUrl"
                  alt="选中图片"
                  class="big-image"
              />
            </div>
          </div>
        </el-card>
        <el-card class="buttons-card">
          <template #header>
            <div class="buttons-header">操作</div>
          </template>
          <div class="buttons">
            <el-button type="primary" @click="startTimer">开始</el-button>
            <el-button type="primary" @click="pauseTimer">暂停</el-button>
            <el-button type="primary" @click="submitData">提交</el-button>
          </div>
        </el-card>
        <el-card class="text-area-card">
          <template #header>
            <div class="text-area-header">文字描述</div>
          </template>
          <el-input
              type="textarea"
              v-model="textDescription"
              placeholder="可编辑的文字描述"
              class="text-area"
              :autosize="{ minRows: 8, maxRows: 20 }"
          />
        </el-card>
      </el-main>

      <!-- 右侧面板使用响应式宽度 -->
      <el-aside class="right-panel">
        <el-radio-group v-model="selectedImageId" class="global-radio-group">
          <el-collapse v-model="activeNames">
            <el-collapse-item
                v-for="(category, catId) in currentPatientCategories"
                :key="catId"
                :title="`类别 ${catId}`"
                :name="`category-${catId}`"
                :style="{ width: '100%' }"
            >
              <div class="category-images-scroll">
                <!-- 使用 Flex 布局实现多行排列 -->
                <div class="category-images flex-container">
                  <ImageCheckboxSelector
                      v-for="img in category.images"
                      :key="img.imageId"
                      :image="img"
                      :selectedImages="selectedImages"
                      @toggle-selection="toggleImageSelection"
                      @clear-selection="handleClearSelection"
                      class="category-image flex-item"
                  />
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </el-radio-group>


        <el-card class="selected-opt-card">
          <template #header>
            <div class="selected-opt-card">选中操作</div>
          </template>
          <el-button type="primary" @click="confirmCategories">确认选择</el-button>
          <el-button
              type="danger"
              @click="clearAllSelections"
              class="clear-button"
          >清除选中</el-button>
        </el-card>
      </el-aside>
    </el-container>

    <div v-if="wsStatus" class="ws-status">{{ wsStatus }}</div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { Clock } from '@element-plus/icons-vue';
import {
  ElButton,
  ElInput,
  ElUpload,
  ElCollapse,
  ElCollapseItem,
  ElRow,
  ElCol,
  ElRadioGroup,
  ElMessageBox
} from 'element-plus';
import ImageCheckboxSelector from '../components/ImageCheckboxSelector.vue';
import 'element-plus/dist/index.css';

import { onMounted, onUnmounted } from 'vue';

// WebSocket 连接
const ws = ref(null);
const wsStatus = ref('');
const currentPatientId = ref(''); // 当前检索的病人 ID

// 病人数据结构：{ [patientId]: { categories: { [catId]: { images: [] } } }
const patients = ref({});
const currentPatientCategories = ref({}); // 当前病人的分类数据

// 生成唯一图片ID（分类ID+序号）
const generateUniqueId = (categoryId, index) => `${categoryId}-${index + 1}`;
// 清除所有选中状态
const clearAllSelections = () => {
  selectedImages.value = [];
  selectedImageId.value = null;
  ElMessageBox.alert('已清除所有选中内容', '操作完成', {
    type: 'info'
  });
};

// 处理清除选择事件，通知子组件更新状态
const handleClearSelection = () => {
  selectedImages.value = [];
};

const time = ref(0);
const isRunning = ref(false);
const selectedImages = ref([]);
const textDescription = ref('检查所见：\n' +
    '\n' +
    '-食道：粘膜光滑柔软，血管纹理清晰，扩张度好，齿状线清晰。\n' +
    '-贲门：未见异常。\n' +
    '-胃底：黏液糊量中，清，粘膜光滑。\n' +
    '-胃体：粘膜光滑，红白相间，蠕动可。\n' +
    '-胃角：弧度存在，粘膜光滑柔软，蠕动可。\n' +
    '-胃窦：粘膜光滑，红白相间，散见充血斑。\n' +
    '-幽门：可见，呈圆形，开闭自如，粘膜光滑柔软。\n' +
    '-十二指肠：球部变形，前壁可见一枚溃疡面，大小约0.8×1.0cm，表面覆白苔，基底少许血痂，周围粘膜充血水肿，降部未见异常。\n' +
    '\n' +
    '诊断结论：1.慢性非萎缩性胃炎，2.十二指肠球部溃疡（活动期）');
const activeNames = ref([]);
const selectedImageId = ref(null);
let timer = null;

// 初始化 WebSocket 连接
onMounted(() => {
  initWebSocket();
});

onUnmounted(() => {
  if (ws.value) {
    ws.value.close();
  }
});

// 初始化 WebSocket
function initWebSocket() {
  ws.value = new WebSocket('ws://localhost:8080/ws');
  ws.value.onopen = () => {
    wsStatus.value = 'WebSocket 连接已建立';
  };
  ws.value.onmessage = (event) => {
    try {
      const imgMeta = JSON.parse(event.data);
      handleReceivedImage(imgMeta);
    } catch (error) {
      wsStatus.value = `数据解析失败: ${error.message}`;
    }
  };
  ws.value.onerror = (error) => {
    wsStatus.value = `WebSocket 错误: ${error.message}`;
  };
  ws.value.onclose = () => {
    wsStatus.value = 'WebSocket 连接已断开';
    setTimeout(() => initWebSocket(), 5000); // 自动重连
  };
}

// 处理接收到的图片元数据
function handleReceivedImage(meta) {
  // 初始化病人数据结构
  if (!patients.value[meta.patientId]) {
    patients.value[meta.patientId] = { categories: {} };
  }
  // 初始化类别数据结构
  if (!patients.value[meta.patientId].categories[meta.categoryId]) {
    patients.value[meta.patientId].categories[meta.categoryId] = {
      id: meta.categoryId,
      name: `类别 ${meta.categoryId}`,
      images: [],
    };
  }
  // 添加图片到对应类别
  patients.value[meta.patientId].categories[meta.categoryId].images.push({
    imageId: meta.imageId,
    url: `data:image/${getMimeType(meta.fileName)};base64,${meta.imageData}`,
    fileName: meta.fileName,
    patientId: meta.patientId, // 保存病人 ID
    categoryId: meta.categoryId, // 保存类别 ID
  });

  if (currentPatientId.value === meta.patientId) {
    currentPatientCategories.value = patients.value[currentPatientId.value].categories;
  }
}

// 检索病人图片
function fetchPatientImages() {
  if (!currentPatientId.value) {
    ElMessageBox.warning('请输入病人 ID');
    return;
  }
  ws.value.send(currentPatientId.value); // 发送病人 ID 请求
  currentPatientCategories.value = patients.value[currentPatientId.value]?.categories || {};
}

function getMimeType(filename) {
  const ext = filename.split('.').pop().toLowerCase();
  switch (ext) {
    case 'jpg', 'jpeg':
      return 'jpeg';
    case 'png':
      return 'png';
    case 'gif':
      return 'gif';
    default:
      return 'image';
  }
}

// 切换图片选择状态
const toggleImageSelection = (image) => {
  const selectedImg = {
    url: image.url,
    patientId: image.patientId,
    categoryId: image.categoryId,
    imageId: image.imageId,
  };

  const index = selectedImages.value.findIndex((img) => img.url === selectedImg.url);
  if (index > -1) {
    selectedImages.value.splice(index, 1);
  } else {
    selectedImages.value.push(selectedImg);
  }
};
// 处理文件上传（生成唯一ID）
const handleFileUpload = (file) => {
  const categoryId = categories.value[0].id; // 默认上传到第一个分类
  const index = categories.value[0].images.length;
  const imgId = generateUniqueId(categoryId, index);
  const imgUrl = URL.createObjectURL(file);
  categories.value[0].images.push({ id: imgId, url: imgUrl });
  selectedImages.value.push(imgUrl);
};

// 开始计时器
const startTimer = () => {
  if (!isRunning.value) {
    timer = setInterval(() => {
      time.value++;
    }, 1000);
    isRunning.value = true;
  }
};

// 暂停计时器
const pauseTimer = () => {
  if (isRunning.value) {
    clearInterval(timer);
    isRunning.value = false;
  }
};

// 提交数据
// 提交数据（包含病人 ID、图片类别、文本描述）
const submitData = () => {
  const submission = {
    textDescription: textDescription.value,
    images: selectedImages.value.map((img) => ({
      url: img.url,
      patientId: img.patientId,
      categoryId: img.categoryId,
      imageId: img.imageId, // 可选：如需唯一标识
    })),
  };

  console.log('提交的数据：', submission);
};




// 确认分类
const confirmCategories = () => {
  console.log('确认分类，选中的图片：', selectedImages.value);
};
</script>

<style scoped>
@import '../assets/scss/HomeView.scss';
</style>