<template>
  <div class="settings-container">
    <div class="settings-sidebar">
      <h2>Settings</h2>
      <ul>
        <li v-for="tab in tabs" :key="tab.id" :class="{ active: currentTab === tab.id }" @click="currentTab = tab.id">
          {{ tab.label }}
        </li>
      </ul>
    </div>

    <div class="settings-content">
      
      <div v-if="currentTab === 'edit-profile'" class="tab-content">
        <h3>Edit Profile</h3>
        <div class="profile-pic-section">
          <img :src="profileForm.profile_picture_url || '/default-avatar.png'" alt="Profile" class="avatar-preview" />
          <button @click="triggerFileInput">Change Photo</button>
          <input type="file" ref="fileInput" @change="handleFileChange" style="display: none" accept="image/png, image/jpeg" />
        </div>
        <form @submit.prevent="saveProfile">
          <div class="form-group">
            <label>Name</label>
            <input v-model="profileForm.name" type="text" required />
          </div>
          <div class="form-group">
            <label>Bio (Max 150 chars)</label>
            <textarea v-model="profileForm.bio" maxlength="150"></textarea>
            <small>{{ (profileForm.bio || '').length }}/150</small>
          </div>
          <div class="form-group">
            <label>Gender</label>
            <select v-model="profileForm.gender">
              <option value="Male">Male</option>
              <option value="Female">Female</option>
              <option value="Prefer not to say">Prefer not to say</option>
            </select>
          </div>
          <button type="submit" class="btn-primary">Submit</button>
        </form>
      </div>

      <div v-if="currentTab === 'notifications'" class="tab-content">
        <h3>Notifications</h3>
        <div class="toggle-group">
          <label>Push Notifications</label>
          <input type="checkbox" v-model="notifSettings.enable_push" @change="saveNotifications" />
        </div>
        <div class="toggle-group">
          <label>Email Notifications</label>
          <input type="checkbox" v-model="notifSettings.enable_email" @change="saveNotifications" />
        </div>
      </div>

      <div v-if="currentTab === 'privacy'" class="tab-content">
        <h3>Account Privacy</h3>
        <div class="toggle-group">
          <div class="label-desc">
            <label>Private Account</label>
            <p>When your account is private, only people you approve can see your photos and videos.</p>
          </div>
          <input type="checkbox" v-model="privacySettings.is_private" @change="savePrivacy" />
        </div>
      </div>

      <div v-if="currentTab === 'close-friends'" class="tab-content">
        <h3>Close Friends</h3>
        <p>We don't send notifications when you edit your close friends list.</p>
        
        <input v-model="searchQuery" @input="searchUsers" placeholder="Search users to add..." class="search-bar"/>
        <div v-if="searchResults.length > 0" class="user-list search-results">
            <div v-for="user in searchResults" :key="user.id" class="user-item">
                <div class="user-info">
                  <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small"/>
                  <span>{{ user.username }}</span>
                </div>
                <button @click="addToCloseFriends(user)">Add</button>
            </div>
        </div>

        <h4>Your List</h4>
        <div class="user-list">
          <div v-for="user in closeFriendsList" :key="user.user_id" class="user-item">
            <div class="user-info">
              <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small" />
              <span>{{ user.username }}</span>
            </div>
            <button @click="removeFromCloseFriends(user.user_id)" class="btn-danger">Remove</button>
          </div>
        </div>
      </div>

      <div v-if="currentTab === 'blocked'" class="tab-content">
        <h3>Blocked Accounts</h3>
        <div class="user-list">
          <div v-for="user in blockedList" :key="user.user_id" class="user-item">
              <div class="user-info">
              <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small" />
              <span>{{ user.username }}</span>
            </div>
            <button @click="unblockUser(user.user_id)" class="btn-secondary">Unblock</button>
          </div>
        </div>
      </div>

      <div v-if="currentTab === 'hide-story'" class="tab-content">
        <h3>Hide Story From</h3>
        
        <input v-model="searchQueryStory" @input="searchUsersStory" placeholder="Search users to hide..." class="search-bar"/>
        <div v-if="searchResultsStory.length > 0" class="user-list search-results">
            <div v-for="user in searchResultsStory" :key="user.id" class="user-item">
                <div class="user-info">
                  <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small"/>
                  <span>{{ user.username }}</span>
                </div>
                <button @click="hideStory(user)">Hide</button>
            </div>
        </div>

        <h4>Hidden Users</h4>
        <div class="user-list">
          <div v-for="user in hiddenStoryList" :key="user.user_id" class="user-item">
              <div class="user-info">
              <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small" />
              <span>{{ user.username }}</span>
            </div>
            <button @click="unhideStory(user.user_id)" class="btn-secondary">Unhide</button>
          </div>
        </div>
      </div>

      <div v-if="currentTab === 'request-verified'" class="tab-content">
        <h3>Request Verification</h3>
        <form @submit.prevent="submitVerification">
          <div class="form-group">
            <label>National Identity Card Number</label>
            <input v-model="verifyForm.national_id" type="text" required placeholder="e.g. 1234567890" />
          </div>
          <div class="form-group">
            <label>Reason for Verification</label>
            <textarea v-model="verifyForm.reason" required placeholder="Why should you be verified?"></textarea>
          </div>
          <div class="form-group">
            <label>Photo of your face (Selfie)</label>
            <input type="file" ref="selfieInput" required accept="image/*" />
          </div>
          <button type="submit" class="btn-primary">Submit Request</button>
        </form>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { settingsApi, usersApi, postsApi } from '@/services/apiService';
import { useAuth } from '@/composables/useAuth';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { user } = useAuth();
const currentTab = ref('edit-profile');

const tabs = [
  { id: 'edit-profile', label: 'Edit Profile' },
  { id: 'notifications', label: 'Notifications' },
  { id: 'privacy', label: 'Account Privacy' },
  { id: 'close-friends', label: 'Close Friends' },
  { id: 'blocked', label: 'Blocked' },
  { id: 'hide-story', label: 'Hide Story' },
  { id: 'request-verified', label: 'Request Verified' },
];

// 1. Edit Profile Logic
const fileInput = ref<HTMLInputElement | null>(null);
const profileForm = reactive({
  name: '',
  bio: '',
  gender: '',
  profile_picture_url: ''
});

const triggerFileInput = () => fileInput.value?.click();

const handleFileChange = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0];
  if (file) {
    try {
      // 1. Get Presigned URL
      const res = await postsApi.generateUploadUrl(file.name, file.type);
      const uploadUrl = res.data.upload_url;
      // 2. Upload to MinIO
      await postsApi.uploadFileToMinio(uploadUrl, file);
      // 3. Set Public URL
      profileForm.profile_picture_url = res.data.public_url || uploadUrl.split('?')[0]; 
    } catch (error) {
      alert("Failed to upload image");
      console.error(error);
    }
  }
};

const saveProfile = async () => {
    try {
        await settingsApi.updateProfile(profileForm);
        alert('Profile saved!');
    } catch(e) { alert('Failed to save profile'); }
};

// 2 & 3. Settings (Notif & Privacy)
const notifSettings = reactive({ enable_push: true, enable_email: true });
const privacySettings = reactive({ is_private: false });

const saveNotifications = async () => {
  try {
    await settingsApi.updateNotifications(notifSettings.enable_push, notifSettings.enable_email);
  } catch (e) { console.error("Failed to update notifications", e); }
};
const savePrivacy = async () => {
  try {
    await settingsApi.updatePrivacy(privacySettings.is_private);
  } catch (e) { console.error("Failed to update privacy", e); }
};

// 4. Close Friends
const closeFriendsList = ref<any[]>([]);
const searchQuery = ref('');
const searchResults = ref<any[]>([]);

const searchUsers = async () => {
    if(searchQuery.value.length < 3) return;
    try {
      const res = await usersApi.searchUsers(searchQuery.value);
      searchResults.value = res.data.users || [];
    } catch(e) { console.error(e); }
};
const addToCloseFriends = async (user: any) => {
    try {
      const id = user.user_id || user.id; 
      await settingsApi.addCloseFriend(id);
      closeFriendsList.value.push(user);
      searchResults.value = [];
      searchQuery.value = '';
    } catch(e) { alert("Failed to add user"); }
};
const removeFromCloseFriends = async (id: string) => {
    try {
      await settingsApi.removeCloseFriend(id);
      closeFriendsList.value = closeFriendsList.value.filter(u => u.user_id !== id);
    } catch(e) { alert("Failed to remove user"); }
};

// 5. Blocked
const blockedList = ref<any[]>([]);
const unblockUser = async (id: string) => {
    try {
      await settingsApi.unblockUser(id);
      blockedList.value = blockedList.value.filter(u => u.user_id !== id);
    } catch(e) { alert("Failed to unblock user"); }
};

// 6. Hide Story
const hiddenStoryList = ref<any[]>([]);
const searchQueryStory = ref('');
const searchResultsStory = ref<any[]>([]);

const searchUsersStory = async () => {
    if(searchQueryStory.value.length < 3) return;
    try {
      const res = await usersApi.searchUsers(searchQueryStory.value);
      searchResultsStory.value = res.data.users || [];
    } catch(e) { console.error(e); }
};
const hideStory = async (user: any) => {
    try {
      await settingsApi.hideStoryFromUser(user.id);
      hiddenStoryList.value.push(user);
      searchResultsStory.value = [];
      searchQueryStory.value = '';
    } catch(e) { alert("Failed to hide story from user"); }
};
const unhideStory = async (id: string) => {
    try {
      await settingsApi.unhideStoryFromUser(id);
      hiddenStoryList.value = hiddenStoryList.value.filter(u => u.user_id !== id);
    } catch(e) { alert("Failed to unhide story"); }
};

// 7. Request Verified
const selfieInput = ref<HTMLInputElement | null>(null);
const verifyForm = reactive({ national_id: '', reason: '' });

const submitVerification = async () => {
    const file = selfieInput.value?.files?.[0];
    if(!file) return alert("Selfie photo required");
    
    try {
      // 1. Upload Image First
      const res = await postsApi.generateUploadUrl(file.name, file.type);
      const uploadUrl = res.data.upload_url;
      await postsApi.uploadFileToMinio(uploadUrl, file);
      const selfieUrl = res.data.public_url || uploadUrl.split('?')[0];

      // 2. Submit Request as JSON (Backend expects JSON binding)
      const requestData = {
          national_id: verifyForm.national_id,
          reason: verifyForm.reason,
          selfie_url: selfieUrl
      };
      
      await settingsApi.requestVerification(requestData);
      
      alert("Request Submitted Successfully!");
      verifyForm.national_id = '';
      verifyForm.reason = '';
      if (selfieInput.value) selfieInput.value.value = ''; // Reset file input
    } catch(e) { 
      alert("Submission failed"); 
      console.error(e);
    }
};

onMounted(async () => {
    try {
      // FIX: Handle potential null values to prevent runtime crashes
      const me = await usersApi.getMe();
      Object.assign(profileForm, {
          name: me.data.name || '',
          bio: me.data.bio || '', // Ensure bio is string
          gender: me.data.gender || 'Prefer not to say',
          profile_picture_url: me.data.profile_picture_url || ''
      });

      // Fetch Preferences
      const prefs = await settingsApi.getSettings();
      if(prefs.data) {
        notifSettings.enable_push = !!prefs.data.enable_push;
        notifSettings.enable_email = !!prefs.data.enable_email;
        privacySettings.is_private = !!prefs.data.is_private;
      }

      // Fetch Lists
      const [cfRes, blRes, hsRes] = await Promise.all([
          settingsApi.getCloseFriends(),
          settingsApi.getBlockedUsers(),
          settingsApi.getHiddenStoryUsers()
      ]);
      
      closeFriendsList.value = cfRes.data.users || [];
      blockedList.value = blRes.data.users || [];
      hiddenStoryList.value = hsRes.data.users || [];
    } catch (e) {
      console.error("Error loading settings data", e);
    }
});
</script>

<style scoped>
.settings-container { 
    display: flex; 
    height: 100vh; 
    background-color: var(--bg-color); 
    color: var(--text-color); 
}

.settings-sidebar { 
    width: 250px; 
    border-right: 1px solid var(--border-color); 
    padding: 20px; 
}

.settings-sidebar li { 
    padding: 15px; cursor: pointer; 
    list-style: none; 
    border-radius: 8px; 
    margin-bottom: 5px;
}


.settings-sidebar li:hover, .settings-sidebar li.active { 
    background-color: var(--hover-color); 
    font-weight: bold; 
}

.settings-content { 
    flex: 1; 
    padding: 40px; 
    overflow-y: auto; 
}

.tab-content { 
    max-width: 600px; 
}

.form-group { 
    margin-bottom: 20px; 
    display: flex; 
    flex-direction: column; 
    gap: 8px;
}

.form-group input, .form-group textarea, .form-group select { 
    padding: 10px; 
    background: var(--input-bg); 
    border: 1px solid var(--border-color); 
    color: white; 
    border-radius: 4px; 
}

.toggle-group { 
    display: flex; 
    justify-content: space-between; 
    align-items: center; 
    margin-bottom: 20px; 
    padding: 10px 0;
    border-bottom: 1px solid var(--border-color);
}
.label-desc {
  max-width: 80%;
}

.user-list { 
    margin-top: 20px; 
}

.user-item { 
    display: flex; 
    align-items: center; 
    justify-content: space-between; 
    margin-bottom: 15px; 
    padding: 8px;
    background: var(--card-bg);
    border-radius: 8px;
}

.user-info { 
    display: flex; 
    align-items: center; 
    gap: 10px; 
}

.avatar-small { 
    width: 40px; 
    height: 40px; 
    border-radius: 50%; 
    object-fit: cover;
}

.avatar-preview {
    width: 100px;
    height: 100px;
    border-radius: 50%;
    object-fit: cover;
    margin-bottom: 10px;
}

.profile-pic-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 20px;
}

.btn-primary { 
    background: #0095f6; 
    color: white; 
    border: none; 
    padding: 10px 20px; 
    border-radius: 4px; 
    cursor: pointer; 
}

.btn-secondary {
    background: #efefef;
    color: black;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer; 
}

.btn-danger { 
    background: #ed4956; 
    color: white; 
    border: none; 
    padding: 8px 16px; 
    border-radius: 4px; 
    cursor: pointer; 
}

.search-bar { 
    width: 100%; 
    padding: 10px; 
    margin-bottom: 10px; 
    background: var(--input-bg); 
    border: 1px solid var(--border-color); 
    color: white; 
    border-radius: 4px;
}

.search-results {
    background: var(--dropdown-bg);
    padding: 10px;
    border-radius: 8px;
    max-height: 200px;
    overflow-y: auto;
}
</style>