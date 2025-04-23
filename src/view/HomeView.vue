<template>
  <div class="app-container">
    <el-container>
      <el-main class="left-panel">
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
                v-for="category in categories"
                :key="category.id"
                :title="category.name"
                :name="`category-${category.id}`"
                :style="{ width: '100%' }"
            >
            <div class="category-images-scroll">
              <div class="category-images">
                <ImageCheckboxSelector
                    v-for="img in category.images"
                    :key="img.id"
                    :image="img"
                    :selectedImages="selectedImages"
                    @toggle-selection="toggleImageSelection"
                    @clear-selection="handleClearSelection"
                    class="category-image"
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
// 初始化数据（唯一ID+两列数据）
const categories = ref([
  {
    id: 1,
    name: 'label1',
    images: [
      { id: generateUniqueId(1, 0), url: 'https://picsum.photos/200/200?random=1' },
      { id: generateUniqueId(1, 1), url: 'https://picsum.photos/200/200?random=2' },
      { id: generateUniqueId(1, 2), url: 'https://picsum.photos/200/200?random=3' }
    ]
  },
  {
    id: 2,
    name: 'label2',
    images: [
      { id: generateUniqueId(2, 0), url: 'https://picsum.photos/200/200?random=4' },
      { id: generateUniqueId(2, 1), url: 'https://picsum.photos/200/200?random=5' },
      { id: generateUniqueId(2, 2), url: 'https://picsum.photos/200/200?random=6' }
    ]
  },
  {
    id: 3,
    name: 'label3',
    images: [
      { id: generateUniqueId(3, 0), url: 'https://picsum.photos/200/200?random=7' },
      { id: generateUniqueId(3, 1), url: 'https://picsum.photos/200/200?random=8' },
      { id: generateUniqueId(3, 2), url: 'https://picsum.photos/200/200?random=9' }
    ]
  },
  // 新增第4组
  {
    id: 4,
    name: 'label4',
    images: [
      { id: generateUniqueId(4, 0), url: 'https://picsum.photos/200/200?random=10' },
      { id: generateUniqueId(4, 1), url: 'https://picsum.photos/200/200?random=11' },
      { id: generateUniqueId(4, 2), url: 'https://picsum.photos/200/200?random=12' }
    ]
  },
  // 新增第5组
  {
    id: 5,
    name: 'label5',
    images: [
      { id: generateUniqueId(5, 0), url: 'https://picsum.photos/200/200?random=13' },
      { id: generateUniqueId(5, 1), url: 'https://picsum.photos/200/200?random=14' },
      { id: generateUniqueId(5, 2), url: 'https://picsum.photos/200/200?random=15' }
    ]
  },
  // 新增第6组
  {
    id: 6,
    name: 'label6',
    images: [
      { id: generateUniqueId(6, 0), url: 'https://picsum.photos/200/200?random=16' },
      { id: generateUniqueId(6, 1), url: 'https://picsum.photos/200/200?random=17' },
      { id: generateUniqueId(6, 2), url: 'https://picsum.photos/200/200?random=18' }
    ]
  },
  // 新增第7组
  {
    id: 7,
    name: 'label7',
    images: [
      { id: generateUniqueId(7, 0), url: 'https://picsum.photos/200/200?random=19' },
      { id: generateUniqueId(7, 1), url: 'https://picsum.photos/200/200?random=20' },
      { id: generateUniqueId(7, 2), url: 'https://picsum.photos/200/200?random=21' },
      { id: generateUniqueId(7, 3), url: 'https://picsum.photos/200/200?random=22' },
      { id: generateUniqueId(7, 4), url: 'https://picsum.photos/200/200?random=23' },
      { id: generateUniqueId(7, 5), url: 'https://picsum.photos/200/200?random=24' },
      { id: generateUniqueId(7, 6), url: 'https://picsum.photos/200/200?random=25' },
      { id: generateUniqueId(7, 7), url: 'https://picsum.photos/200/200?random=26' },
      { id: generateUniqueId(7, 8), url: 'https://picsum.photos/200/200?random=27' },
    ]
  }
]);
const time = ref(0);
const isRunning = ref(false);
const selectedImages = ref([]);
const textDescription = ref('');
const activeNames = ref([]);
const selectedImageId = ref(null);
let timer = null;

// 切换图片选择状态
const toggleImageSelection = (image) => {
  const index = selectedImages.value.findIndex((imgUrl) => imgUrl === image.url);
  if (index > -1) {
    selectedImages.value.splice(index, 1);
  } else {
    selectedImages.value.push(image.url);
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
const submitData = () => {
  console.log('提交的数据：', {
    selectedImages: selectedImages.value,
    textDescription: textDescription.value
  });
};



// 确认分类
const confirmCategories = () => {
  console.log('确认分类，选中的图片：', selectedImages.value);
};
</script>

<style scoped>
@import '../assets/scss/HomeView.scss';
</style>