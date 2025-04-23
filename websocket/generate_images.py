import os
import random
import requests
from faker import Faker

# 配置参数
NUM_PATIENTS = 10         # 生成的病人数量
IMAGES_PER_CATEGORY = 5  # 每个类别下的图片数量
CATEGORIES = 7           # 每个病人的类别数（1-7）
IMAGE_EXTENSIONS = ['jpg', 'png', 'gif']  # 支持的图片扩展名
BASE_URL = "https://picsum.photos/300/300?random="  # 随机图片源

# 目录结构：patient_{id}/category_{n}/img_{idx}.ext
def create_patient_structure(root_dir):
    fake = Faker()

    # 创建根目录
    os.makedirs(root_dir, exist_ok=True)

    for patient_id in range(1, NUM_PATIENTS + 1):
        patient_dir = os.path.join(root_dir, f"patient_{patient_id}")
        os.makedirs(patient_dir, exist_ok=True)

        for category_id in range(1, CATEGORIES + 1):
            category_dir = os.path.join(patient_dir, f"category_{category_id}")
            os.makedirs(category_dir, exist_ok=True)

            # 生成随机图片
            for idx in range(1, IMAGES_PER_CATEGORY + 1):
                ext = random.choice(IMAGE_EXTENSIONS)
                img_name = f"img_{idx:03d}.{ext}"
                img_path = os.path.join(category_dir, img_name)

                # 下载随机图片
                if not os.path.exists(img_path):
                    try:
                        response = requests.get(f"{BASE_URL}{fake.random_number(digits=5)}")
                        response.raise_for_status()

                        with open(img_path, "wb") as f:
                            f.write(response.content)
                        print(f"已生成图片: {img_path}")
                    except Exception as e:
                        print(f"下载图片失败: {e}")

if __name__ == "__main__":
    root_directory = "./patient_images"  # 根目录
    create_patient_structure(root_directory)
    print(f"\n目录结构已生成至: {root_directory}")
    print(f"总病人数: {NUM_PATIENTS}, 每个病人{IMAGES_PER_CATEGORY}张图片/7个类别")