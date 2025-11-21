<template>
  <div v-if="isOpen" class="create-overlay" @click="closeCreate">
    <div class="create-modal" @click.stop>
      <div v-if="selectedFiles.length === 0" class="upload-step">
        <div class="modal-header">
          <h2>Create new post</h2>
          <button class="close-btn" @click="closeCreate">✕</button>
        </div>
        <div class="modal-body">
          <div class="upload-area" @dragover.prevent="dragOver = true" @dragleave="dragOver = false" @drop.prevent="handleDrop">
            <div :class="['upload-content', { dragover: dragOver }]">
              <p class="upload-text">Drag photos and videos here</p>
              <button class="select-btn" @click="triggerFileInput">Select from computer</button>
            </div>
            <input 
              ref="fileInput"
              type="file" 
              multiple 
              accept="image/*,video/*"
              style="display: none"
              @change="handleFileSelect"
            />
          </div>
        </div>
      </div>

      <div v-else class="edit-step">
        <div class="modal-header">
          <button class="back-btn" @click="goBack">← Back</button>
          <h2>Create new post</h2>
          <button 
            class="share-btn" 
            @click="handleSharePost"
            :disabled="isUploading"
          >
            {{ isUploading ? 'Sharing...' : 'Share' }}
          </button>
        </div>
        
        <div class="edit-container">
          <div class="preview-area">
            <div class="media-wrapper" v-if="currentFile">
              <video 
                v-if="currentFile.type.startsWith('video/')" 
                :src="currentPreviewUrl" 
                controls 
                class="preview-image"
              ></video>
              <img 
                v-else 
                :src="currentPreviewUrl" 
                :alt="currentFile.name" 
                class="preview-image" 
              />
            </div>

            <button 
              v-if="selectedFiles.length > 1 && currentIndex > 0" 
              class="nav-btn left" 
              @click="currentIndex--"
            >❮</button>
            
            <button 
              v-if="selectedFiles.length > 1 && currentIndex < selectedFiles.length - 1" 
              class="nav-btn right" 
              @click="currentIndex++"
            >❯</button>
            
            <div class="dots-container" v-if="selectedFiles.length > 1">
              <span 
                v-for="(_, index) in selectedFiles" 
                :key="index" 
                class="dot"
                :class="{ active: index === currentIndex }"
                @click="currentIndex = index"
              ></span>
            </div>
          </div>

          <div class="caption-area">
            <textarea 
              v-model="postDescription"
              placeholder="Write a caption..."
              class="caption-input"
              @input="updateWordCount"
            ></textarea>
            <div class="caption-footer">
              <span class="word-count">{{ wordCount }} / 2200</span>
            </div>

            <input 
              v-model="location"
              type="text"
              placeholder="Add location"
              class="location-input"
            />

            <div v-if="isUploading" class="upload-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
              </div>
              <p class="progress-text">{{ uploadProgress }}% uploaded</p>
            </div>

            <div v-if="errorMessage" class="error-message">
              {{ errorMessage }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import axios from 'axios'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success'): void
}>()

const fileInput = ref<HTMLInputElement>()
const dragOver = ref(false)

// --- UPDATED STATE FOR MULTIPLE FILES ---
const selectedFiles = ref<File[]>([])
const filePreviews = ref<string[]>([])
const currentIndex = ref(0)

const postDescription = ref('')
const location = ref('')
const wordCount = ref(0)
const isUploading = ref(false)
const uploadProgress = ref(0)
const errorMessage = ref('')

// Helper to get current visible file
const currentFile = computed(() => selectedFiles.value[currentIndex.value])
const currentPreviewUrl = computed(() => filePreviews.value[currentIndex.value])

const closeCreate = () => {
  if (!isUploading.value) {
    resetForm()
    emit('close')
  }
}

const goBack = () => {
  if (!isUploading.value) {
    // Only clear files if user goes back from edit step
    resetForm()
  }
}

const resetForm = () => {
  selectedFiles.value = []
  // Revoke old URLs to avoid memory leaks
  filePreviews.value.forEach(url => URL.revokeObjectURL(url))
  filePreviews.value = []
  currentIndex.value = 0
  
  postDescription.value = ''
  location.value = ''
  wordCount.value = 0
  isUploading.value = false
  uploadProgress.value = 0
  errorMessage.value = ''
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const updateWordCount = () => {
  const description = postDescription.value
  const words = description.trim().split(/\s+/).filter(word => word.length > 0)
  wordCount.value = words.length
}

const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    processFiles(Array.from(input.files))
  }
}

const handleDrop = (event: DragEvent) => {
  dragOver.value = false
  const droppedFiles = event.dataTransfer?.files
  if (droppedFiles && droppedFiles.length > 0) {
    processFiles(Array.from(droppedFiles))
  }
}

const processFiles = (files: File[]) => {
  const validFiles = files.filter(file => 
    file.type.startsWith('image/') || file.type.startsWith('video/')
  )
  
  validFiles.forEach(file => {
    selectedFiles.value.push(file)
    filePreviews.value.push(URL.createObjectURL(file))
  })
}

const handleSharePost = async () => {
  if (selectedFiles.value.length === 0) {
    errorMessage.value = 'Please select a file to upload'
    return
  }

  try {
    isUploading.value = true
    errorMessage.value = ''
    uploadProgress.value = 0

    // Get auth token
    const accessToken = localStorage.getItem('accessToken')
    if (!accessToken) {
      errorMessage.value = 'You must be logged in to create a post'
      return
    }

    const mediaObjects = []
    const totalFiles = selectedFiles.value.length
    const progressPerFile = 90 / totalFiles 

    // --- LOOP: UPLOAD EACH FILE ---
    for (let i = 0; i < totalFiles; i++) {
      const file = selectedFiles.value[i]
      
      // FIX: Add this check to satisfy TypeScript
      if (!file) continue 

      // 1. Get Presigned URL
      const urlResponse = await axios.get('/api/v1/posts/generate-upload-url', {
        params: {
          file_name: file.name,
          file_type: file.type
        },
        headers: { 'Authorization': `Bearer ${accessToken}` }
      })

      const { upload_url, object_name } = urlResponse.data

      // 2. Upload to MinIO
      await axios.put(upload_url, file, {
        headers: { 'Content-Type': file.type },
        onUploadProgress: (progressEvent) => {
           // Optional: finer grain progress update
        }
      })

      // 3. Add to list
      mediaObjects.push({
        media_object_name: object_name,
        media_type: file.type
      })
      
      uploadProgress.value = Math.round((i + 1) * progressPerFile)
    }

    // --- FINAL STEP: CREATE POST WITH ARRAY ---
    await axios.post('/api/v1/posts', {
      media: mediaObjects, 
      caption: postDescription.value.trim(),
      location: location.value.trim()
    }, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    })

    uploadProgress.value = 100

    setTimeout(() => {
      resetForm()
      emit('success')
      emit('close')
    }, 500)

  } catch (error: any) {
    console.error('Failed to make post:', error)
    if (error.response) {
      errorMessage.value = error.response.data?.error || 'Failed to create post. Please try again.'
    } else {
      errorMessage.value = 'An unexpected error occurred. Please try again.'
    }
  } finally {
    if (errorMessage.value) {
      isUploading.value = false
      uploadProgress.value = 0
    }
  }
}
</script>

<style scoped>
.create-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  z-index: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.create-modal {
  background: var(--background-dark);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  width: 100%;
  max-width: 750px; /* Increased slightly to accommodate carousel better */
  max-height: 80vh;
  overflow: hidden;
  animation: slideUp 0.3s ease;
  display: flex;
  flex-direction: column;
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px; /* Compact header */
  border-bottom: 1px solid var(--border-color);
  height: 45px;
}

.modal-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  flex: 1;
  text-align: center;
}

.close-btn,
.back-btn {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 20px;
  cursor: pointer;
  padding: 0;
}

.share-btn {
  background: none;
  color: var(--primary-color);
  border: none;
  padding: 0;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
}

.share-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.share-btn:hover:not(:disabled) {
  color: var(--text-primary);
}

/* Upload Step */
.modal-body {
  padding: 40px;
  height: 400px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.upload-area {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-content {
  text-align: center;
  transition: all 0.3s ease;
}

.upload-content.dragover {
  transform: scale(1.05);
}

.upload-text {
  color: var(--text-primary);
  margin: 0 0 20px 0;
  font-size: 20px;
}

.select-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
}

/* Edit Step */
.edit-container {
  display: flex;
  height: 500px; /* Fixed height for split view */
}

.preview-area {
  flex: 1.5; /* Takes up more space */
  background: #000;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.media-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* Carousel Controls */
.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(26, 26, 26, 0.8);
  color: white;
  border: none;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  z-index: 10;
  transition: background 0.2s;
}
.nav-btn:hover { background: rgba(255, 255, 255, 0.2); }
.left { left: 12px; }
.right { right: 12px; }

.dots-container {
  position: absolute;
  bottom: 15px;
  display: flex;
  gap: 6px;
  z-index: 10;
}

.dot {
  width: 6px;
  height: 6px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 50%;
  cursor: pointer;
  transition: background 0.2s;
}
.dot.active {
  background: #fff;
  transform: scale(1.2);
}

/* Caption Area */
.caption-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--border-color);
  padding: 16px;
}

.caption-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  padding: 0;
  font-family: inherit;
  font-size: 14px;
  resize: none;
}

.caption-input:focus {
  outline: none;
}

.caption-footer {
  display: flex;
  justify-content: flex-end;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-color);
}

.word-count {
  font-size: 12px;
  color: var(--text-secondary);
}

.location-input {
  background: transparent;
  border: none;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 12px 0;
  font-size: 14px;
}

.location-input:focus {
  outline: none;
}

.upload-progress {
  margin-top: 20px;
}

.progress-bar {
  height: 2px;
  background: var(--surface-light);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(45deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
  transition: width 0.3s ease;
}

.progress-text {
  margin-top: 5px;
  font-size: 12px;
  color: var(--text-secondary);
}

.error-message {
  color: #ed4956;
  font-size: 12px;
  margin-top: 10px;
  text-align: center;
}

@media (max-width: 768px) {
  .create-modal {
    max-width: 100%;
    height: 100%;
    max-height: 100%;
    border-radius: 0;
  }

  .edit-container {
    flex-direction: column;
    height: auto;
    flex: 1;
  }

  .preview-area {
    height: 300px;
    flex: none;
  }
}
</style>