@ -1,210 +1,207 @@
<template>
  <div class="saved-container">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
    </div>

    <div v-else-if="collections.length === 0" class="empty-state">
      <div class="empty-icon-wrapper">
         <img src="/icons/save-icon.png" class="empty-icon-img" alt="Save"/>
      </div>
      <h3>Start Saving</h3>
      <p>Save photos and videos to your private collections.</p>
    </div>

    <div v-else class="collections-grid">
      <div 
        v-for="col in collections" 
        :key="col.id" 
        class="collection-card"
        @click="handleCollectionClick(col.id)"
      >
        <div class="collection-cover">
          <div class="cover-collage">
            <div 
              class="collage-item" 
              :style="getImageStyle(col.cover_images, 0)"
            ></div>
            <div 
              class="collage-item" 
              :style="getImageStyle(col.cover_images, 1)"
            ></div>
            <div 
              class="collage-item" 
              :style="getImageStyle(col.cover_images, 2)"
            ></div>
            <div 
              class="collage-item" 
              :style="getImageStyle(col.cover_images, 3)"
            ></div>
          </div>
        </div>

        <div class="collection-info">
          <span class="collection-name">{{ col.name }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { postsApi } from '../services/apiService';

const collections = ref<any[]>([]);
const loading = ref(true);
const router = useRouter();

const fetchCollections = async () => {
  try {
    const res = await postsApi.getUserCollections();
    collections.value = res.data.collections || res.data || [];
  } catch (error) {
    console.error("Failed to fetch collections", error);
  } finally {
    loading.value = false;
  }
};

const getImageStyle = (images: string[], index: number) => {
  if (images && images[index]) {
    return { 
      backgroundImage: `url(${images[index]})`,
      backgroundColor: '#262626' 
    };
  }
  return { backgroundColor: '#262626' };
};

const handleCollectionClick = (id: string) => {
  const collection = collections.value.find(c => c.id === id);
  const name = collection ? collection.name : 'Collection';
  
  router.push({ 
    name: 'collection-detail', 
    params: { collectionID: id },
    query: { name: name } 
  });
};

onMounted(() => {
  fetchCollections();
});
</script>

<style scoped>
.saved-container {
  width: 100%;
  padding-top: 15px;
}

.collections-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr); 
  gap: 15px; 
  width: 100%;
}

.collection-card {
  cursor: pointer;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px; /* Space between image and text */
}

/* Cover is strictly square */
.collection-cover {
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: 6px; 
  overflow: hidden;
  border: 1px solid #363636; 
  position: relative;
  transition: opacity 0.2s;
}

.collection-card:hover .collection-cover {
  opacity: 0.85;
}

.cover-collage {
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 0; /* No gap inside = Bigger, cleaner look */
  background: #262626;
}

.collage-item {
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.collection-info {
  padding: 0 2px;
}

.collection-name {
  color: white;
  font-size: 18px; /* Larger font */
  font-weight: 700; /* Bolder text */
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: 0.3px;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 100px 20px;
  color: #a0a0a0;
}

.empty-icon-wrapper {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 2px solid #555;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}

.empty-icon-img {
  width: 36px;
  height: 36px;
  filter: invert(0.8);
}

/* Loading */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px;
}

.spinner {
  width: 30px;
  height: 30px;
  border: 2px solid #333;
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}


@media (max-width: 600px) {
  .collections-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }
}
</style>