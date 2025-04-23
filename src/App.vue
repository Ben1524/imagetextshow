<template>
  <HomeView/>
</template>

<script setup>
import { ref } from 'vue';
import { ElSelect, ElOption, ElUpload, ElButton, ElRadioGroup } from 'element-plus';
import ImageRadioSelector from './components/ImageCheckboxSelector.vue';
import 'element-plus/dist/index.css';
import HomeView from "./view/HomeView.vue";

// 初始化 WebSocket 连接（这里假设已经连接好，可参考之前代码完善）
const ws = ref(null);
// 存储所有图片数据
const images = ref([]);
// 存储分类数据
const categories = ref([]);
// 当前选中的分类
const selectedCategory = ref('');
// 过滤后的图片数据
const filteredImages = ref([]);
// 当前选中的图片 ID
const selectedImageId = ref(null);

// 模拟 WebSocket 接收数据
const mockWebSocketData = () => {
  // 模拟分类数据
  categories.value = [
    { id: 1, name: '风景' },
    { id: 2, name: '人物' }
  ];
  // 模拟图片数据
  images.value = [
    { id: 1, url: 'https://picsum.photos/200/200?random=1', categoryId: 1 },
    { id: 2, url: 'https://picsum.photos/200/200?random=2', categoryId: 2 },
    { id: 3, url: 'https://picsum.photos/200/200?random=3', categoryId: 1 }
  ];
  filterImages();
};

// 过滤图片
const filterImages = () => {
  if (selectedCategory.value === '') {
    filteredImages.value = images.value;
  } else {
    filteredImages.value = images.value.filter(
        (image) => image.categoryId === selectedCategory.value
    );
  }
};

// 获取分类名称
const getCategoryName = (categoryId) => {
  const category = categories.value.find((cat) => cat.id === categoryId);
  return category ? category.name : '未分类';
};

// 处理文件选择
const handleFileChange = (file) => {
  const formData = new FormData();
  formData.append('image', file.raw);
  const categoryId = categories.value.length > 0 ? categories.value[0].id : null;
  if (categoryId) {
    formData.append('categoryId', categoryId);
  }

  // 模拟发送文件到后端
  // 这里可以添加实际的 WebSocket 发送逻辑
  console.log('模拟上传文件:', formData);
};

// 切换图片选中状态
const toggleImageSelection = (imageId) => {
  selectedImageId.value = imageId;
};

// 初始化数据
mockWebSocketData();
</script>

<style scoped>
.app-container {
  padding: 20px;
}

.thumbnail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 20px;
  margin-top: 20px;
}
</style>