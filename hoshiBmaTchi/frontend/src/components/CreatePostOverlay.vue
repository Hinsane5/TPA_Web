<template>
  <div v-if="isOpen" class="create-overlay" @click="closeCreate">
    <div class="create-modal" @click.stop>
      <!-- Step 1: Upload -->
      <div v-if="!selectedFile" class="upload-step">
        <div class="modal-header">
          <h2>Create new post</h2>
          <button class="close-btn" @click="closeCreate">✕</button>
        </div>
        <div class="modal-body">
          <div class="upload-area" @dragover.prevent="dragOver = true" @dragleave="dragOver = false" @drop.prevent="handleDrop">
            <div :class="['upload-content', { dragover: dragOver }]">
              <div class="upload-icon">
                <!-- <img src="/icons/upload-placeholder.png" alt="Upload" /> -->
              </div>
              <p class="upload-text">Drag photos and videos here</p>
              <button class="select-btn" @click="triggerFileInput">Select from computer</button>
            </div>
            <input 
              ref="fileInput"
              type="file" 
              accept="image/*,video/*"
              style="display: none"
              @change="handleFileSelect"
            />
          </div>
        </div>
      </div>

      <!-- Step 2: Edit & Caption -->
      <div v-else class="edit-step">
        <div class="modal-header">
          <button class="back-btn" @click="goBack">
            ← Back
          </button>
          <h2>Create new post</h2>
          <button class="share-btn">Share</button>
        </div>
        
        <div class="edit-container">
          <!-- Image Preview -->
          <div class="preview-area">
            <img v-if="filePreview" :src="filePreview" :alt="selectedFile?.name" class="preview-image" />
          </div>

          <!-- Caption Area -->
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

            <!-- Add Location -->
            <button class="add-option">
              📍 Add location
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  close: []
  upload: [file: File, description: string]
}>()

const fileInput = ref<HTMLInputElement>()
const dragOver = ref(false)
const selectedFile = ref<File | null>(null)
const filePreview = ref<string>('')
const postDescription = ref('')
const wordCount = ref(0)

const closeCreate = () => {
  resetForm()
  emit('close')
}

const goBack = () => {
  selectedFile.value = null
  filePreview.value = ''
  postDescription.value = ''
  wordCount.value = 0
}

const resetForm = () => {
  selectedFile.value = null
  filePreview.value = ''
  postDescription.value = ''
  wordCount.value = 0
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const updateWordCount = () => {
  const description = postDescription.value
  wordCount.value = description.trim().split(/\s+/).filter(word => word.length > 0).length
}

const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files?.[0]) {
    selectFile(input.files[0])
  }
}

const handleDrop = (event: DragEvent) => {
  dragOver.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file && (file.type.startsWith('image/') || file.type.startsWith('video/'))) {
    selectFile(file)
  }
}

const selectFile = (file: File) => {
  selectedFile.value = file
  
  // Create preview
  const reader = new FileReader()
  reader.onload = (e) => {
    filePreview.value = e.target?.result as string
  }
  reader.readAsDataURL(file)
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
  max-width: 600px;
  max-height: 80vh;
  overflow: hidden;
  animation: slideUp 0.3s ease;
  display: flex;
  flex-direction: column;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
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
  padding: 5px 10px;
}

.share-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
}

.share-btn:hover {
  background: var(--primary-hover);
}

/* Upload Step */
.modal-body {
  padding: 40px;
  flex: 1;
  overflow-y: auto;
}

.upload-area {
  border: 2px dashed var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.upload-content {
  padding: 60px 20px;
  text-align: center;
  transition: all 0.3s ease;
}

.upload-content.dragover {
  background: var(--surface-light);
  border-color: var(--primary-color);
}

.upload-icon {
  font-size: 40px;
  margin-bottom: 15px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-icon img {
  width: 60px;
  height: 60px;
  object-fit: contain;
}

.upload-text {
  color: var(--text-primary);
  margin: 0 0 20px 0;
  font-size: 16px;
}

.select-btn {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 10px 24px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: background 0.2s ease;
}

.select-btn:hover {
  background: var(--primary-hover);
}

/* Edit Step */
.edit-container {
  display: flex;
  height: 400px;
  gap: 20px;
  padding: 20px;
}

.preview-area {
  flex: 1;
  background: var(--surface-light);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.caption-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.caption-input {
  flex: 1;
  background: var(--surface-light);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 12px;
  border-radius: 8px;
  font-family: inherit;
  font-size: 14px;
  resize: none;
}

.caption-input::placeholder {
  color: var(--text-secondary);
}

.caption-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2px;
}

.word-count {
  font-size: 12px;
  color: var(--text-secondary);
}

.add-option {
  background: none;
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s ease;
}

.add-option:hover {
  background: var(--surface-light);
}

@media (max-width: 768px) {
  .create-modal {
    max-width: 90vw;
    max-height: 90vh;
  }

  .edit-container {
    flex-direction: column;
    height: auto;
    max-height: 60vh;
  }

  .preview-area {
    height: 250px;
  }
}
</style>
