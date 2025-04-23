<template>
  <el-checkbox :label="image" v-model="isSelected" @change="handleChange">
    <img :src="image.url" alt="thumbnail" class="thumbnail-image" />
  </el-checkbox>
</template>

<script setup>
import {ref, watch} from 'vue';
const props = defineProps(['image', 'selectedImages']);
const emit = defineEmits(['toggle-selection']);

const isSelected = ref(props.selectedImages.includes(props.image.url));

const handleChange = () => {
  emit('toggle-selection', props.image);
};

// 监听 selectedImages 变化，更新自身状态
watch(
    () => props.selectedImages,
    (newSelectedImages) => {
      isSelected.value = newSelectedImages.includes(props.image.url);
    },
    { deep: true }
);

</script>

<style scoped>
.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
</style>