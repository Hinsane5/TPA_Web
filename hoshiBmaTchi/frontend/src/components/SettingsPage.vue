<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { settingsApi, usersApi, postsApi } from '@/services/apiService';
import { useAuth } from '@/composables/useAuth';

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
      const res = await postsApi.generateUploadUrl(file.name, file.type);
      const uploadUrl = res.data.upload_url;
      await postsApi.uploadFileToMinio(uploadUrl, file);
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

const blockedList = ref<any[]>([]);
const unblockUser = async (id: string) => {
    try {
      await settingsApi.unblockUser(id);
      blockedList.value = blockedList.value.filter(u => u.user_id !== id);
    } catch(e) { alert("Failed to unblock user"); }
};

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

const selfieInput = ref<HTMLInputElement | null>(null);
const verifyForm = reactive({ national_id: '', reason: '' });

const submitVerification = async () => {
    const file = selfieInput.value?.files?.[0];
    if(!file) return alert("Selfie photo required");
    
    try {
      const res = await postsApi.generateUploadUrl(file.name, file.type);
      const uploadUrl = res.data.upload_url;
      await postsApi.uploadFileToMinio(uploadUrl, file);
      const selfieUrl = res.data.public_url || uploadUrl.split('?')[0];

      const requestData = {
          national_id: verifyForm.national_id,
          reason: verifyForm.reason,
          selfie_url: selfieUrl
      };
      
      await settingsApi.requestVerification(requestData);
      
      alert("Request Submitted Successfully!");
      verifyForm.national_id = '';
      verifyForm.reason = '';
      if (selfieInput.value) selfieInput.value.value = ''; 
    } catch(e) { 
      alert("Submission failed"); 
      console.error(e);
    }
};

onMounted(async () => {
    try {
      const me = await usersApi.getMe();
      Object.assign(profileForm, {
          name: me.data.name || '',
          bio: me.data.bio || '', 
          gender: me.data.gender || 'Prefer not to say',
          profile_picture_url: me.data.profile_picture_url || ''
      });

      const prefs = await settingsApi.getSettings();
      if(prefs.data) {
        notifSettings.enable_push = !!prefs.data.enable_push;
        notifSettings.enable_email = !!prefs.data.enable_email;
        privacySettings.is_private = !!prefs.data.is_private;
      }

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

<template>
  <div class="settings-page">
    <div class="settings-container">
      <aside class="settings-sidebar">
        <h2 class="sidebar-title">Settings</h2>
        <ul class="sidebar-menu">
          <li 
            v-for="tab in tabs" 
            :key="tab.id" 
            :class="{ active: currentTab === tab.id }" 
            @click="currentTab = tab.id"
          >
            {{ tab.label }}
          </li>
        </ul>
      </aside>

      <main class="settings-content">
        
        <div v-if="currentTab === 'edit-profile'" class="content-wrapper edit-profile-view">
          <header class="profile-header">
            <div class="profile-avatar-container">
              <img :src="profileForm.profile_picture_url || '/default-avatar.png'" alt="Profile" class="avatar-large" />
            </div>
            <div class="profile-actions">
              <span class="username-display">{{ profileForm.name || 'User' }}</span>
              <button class="link-btn" @click="triggerFileInput">Change Profile Photo</button>
              <input ref="fileInput" type="file" class="hidden-input" accept="image/png, image/jpeg" @change="handleFileChange" />
            </div>
          </header>

          <form class="ig-form" @submit.prevent="saveProfile">
            <div class="form-row">
              <aside><label>Name</label></aside>
              <div class="field">
                <input v-model="profileForm.name" type="text" placeholder="Name" />
                <p class="help-text">Help people discover your account by using the name you're known by.</p>
              </div>
            </div>

            <div class="form-row">
              <aside><label>Bio</label></aside>
              <div class="field">
                <textarea v-model="profileForm.bio" maxlength="150" rows="3"></textarea>
                <div class="char-count">{{ (profileForm.bio || '').length }} / 150</div>
              </div>
            </div>

            <div class="form-row">
              <aside><label>Gender</label></aside>
              <div class="field">
                <select v-model="profileForm.gender" class="custom-select">
                  <option value="Male">Male</option>
                  <option value="Female">Female</option>
                </select>
              </div>
            </div>

            <div class="form-row submit-row">
              <aside></aside>
              <div class="field">
                <button type="submit" class="btn-primary">Submit</button>
              </div>
            </div>
          </form>
        </div>

        <div v-if="currentTab === 'notifications'" class="content-wrapper">
          <h3 class="section-title">Notifications</h3>
          <div class="setting-item">
            <div class="setting-info">
              <label>Push Notifications</label>
              <p>Receive push notifications on your device.</p>
            </div>
            <label class="toggle-switch">
              <input v-model="notifSettings.enable_push" type="checkbox" @change="saveNotifications" />
              <span class="slider"></span>
            </label>
          </div>
          <div class="setting-item">
            <div class="setting-info">
              <label>Email Notifications</label>
              <p>Receive updates via email.</p>
            </div>
            <label class="toggle-switch">
              <input v-model="notifSettings.enable_email" type="checkbox" @change="saveNotifications" />
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div v-if="currentTab === 'privacy'" class="content-wrapper">
          <h3 class="section-title">Account Privacy</h3>
          <div class="setting-item">
            <div class="setting-info">
              <label>Private Account</label>
              <p>When your account is private, only people you approve can see your photos and videos.</p>
            </div>
            <label class="toggle-switch">
              <input v-model="privacySettings.is_private" type="checkbox" @change="savePrivacy" />
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div v-if="currentTab === 'close-friends'" class="content-wrapper">
          <h3 class="section-title">Close Friends</h3>
          <p class="section-desc">We don't send notifications when you edit your close friends list.</p>
          
          <div class="search-container">
            <input v-model="searchQuery" placeholder="Search" class="ig-input search-input" @input="searchUsers"/>
          </div>

          <div v-if="searchResults.length > 0" class="user-list search-results">
            <div v-for="user in searchResults" :key="user.id" class="user-row">
              <div class="user-left">
                <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small"/>
                <span class="username">{{ user.username }}</span>
              </div>
              <button class="btn-text-blue" @click="addToCloseFriends(user)">Add</button>
            </div>
          </div>

          <h4 v-if="closeFriendsList.length > 0" class="sub-header">Your List</h4>
          <div class="user-list">
            <div v-for="user in closeFriendsList" :key="user.user_id" class="user-row">
              <div class="user-left">
                <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small" />
                <span class="username">{{ user.username }}</span>
              </div>
              <button class="btn-danger-outline" @click="removeFromCloseFriends(user.user_id)">Remove</button>
            </div>
            <div v-if="closeFriendsList.length === 0" class="empty-state">No close friends yet.</div>
          </div>
        </div>

        <div v-if="currentTab === 'blocked'" class="content-wrapper">
          <h3 class="section-title">Blocked Accounts</h3>
          <p class="section-desc">You can block people anytime from their profiles.</p>
          <div class="user-list">
            <div v-for="user in blockedList" :key="user.user_id" class="user-row">
               <div class="user-left">
                <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small" />
                <span class="username">{{ user.username }}</span>
              </div>
              <button class="btn-secondary" @click="unblockUser(user.user_id)">Unblock</button>
            </div>
            <div v-if="blockedList.length === 0" class="empty-state">You haven't blocked anyone.</div>
          </div>
        </div>

        <div v-if="currentTab === 'hide-story'" class="content-wrapper">
          <h3 class="section-title">Hide Story From</h3>
          
          <div class="search-container">
            <input v-model="searchQueryStory" placeholder="Search" class="ig-input search-input" @input="searchUsersStory"/>
          </div>

          <div v-if="searchResultsStory.length > 0" class="user-list search-results">
            <div v-for="user in searchResultsStory" :key="user.id" class="user-row">
              <div class="user-left">
                <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small"/>
                <span class="username">{{ user.username }}</span>
              </div>
              <button class="btn-text-blue" @click="hideStory(user)">Hide</button>
            </div>
          </div>

          <h4 v-if="hiddenStoryList.length > 0" class="sub-header">Hidden Users</h4>
          <div class="user-list">
            <div v-for="user in hiddenStoryList" :key="user.user_id" class="user-row">
               <div class="user-left">
                <img :src="user.profile_picture_url || '/default-avatar.png'" class="avatar-small" />
                <span class="username">{{ user.username }}</span>
              </div>
              <button class="btn-secondary" @click="unhideStory(user.user_id)">Unhide</button>
            </div>
            <div v-if="hiddenStoryList.length === 0" class="empty-state">Not hiding story from anyone.</div>
          </div>
        </div>

        <div v-if="currentTab === 'request-verified'" class="content-wrapper">
          <h3 class="section-title">Request Verification</h3>
          <p class="section-desc">Apply for Instagram Verification.</p>
          <form class="ig-form stack-form" @submit.prevent="submitVerification">
            <div class="form-group">
              <label>National Identity Card Number</label>
              <input v-model="verifyForm.national_id" type="text" class="ig-input" placeholder="e.g. 1234567890" required />
            </div>
            <div class="form-group">
              <label>Reason for Verification</label>
              <textarea v-model="verifyForm.reason" class="ig-input" rows="3" required placeholder="Why should you be verified?"></textarea>
            </div>
            <div class="form-group">
              <label>Photo of your face (Selfie)</label>
              <input ref="selfieInput" type="file" class="file-control" required accept="image/*" />
            </div>
            <button type="submit" class="btn-primary full-width">Submit Request</button>
          </form>
        </div>

      </main>
    </div>
  </div>
</template>

<style scoped>
/* Base Styles */
.settings-page {
  background-color: var(--bg-color); 
  color: var(--text-color);
  min-height: 100vh;
  display: flex;
  justify-content: center;
  padding: 30px 0; /* Changed to 0 horizontal padding to give container full control */
}

/* FIXED WIDTH CONTAINER
  - width: 100% with a max-width keeps it consistent on large screens.
  - min-width: 900px ensures it doesn't shrink awkwardly on mid-sized screens.
  - margin: 0 auto centers it.
*/
.settings-container {
  display: flex;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  
  width: 100%;
  max-width: 1200px; /* Wider fixed max width */
  min-width: 900px;  /* Prevents shrinking too much */
  
  height: 85vh; /* Fixed height */
  overflow: hidden; /* Only inner content scrolls */
  margin: 0 20px; /* Add margin here instead of padding on parent */
}

/* Sidebar - Fixed Width */
.settings-sidebar {
  width: 250px;
  flex-shrink: 0; /* PREVENTS SHRINKING */
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.sidebar-title {
  display: none; 
}

.sidebar-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.sidebar-menu li {
  padding: 16px 30px;
  font-size: 16px;
  cursor: pointer;
  border-left: 2px solid transparent;
  transition: background-color 0.2s;
}

.sidebar-menu li:hover {
  background-color: var(--hover-color);
  border-left-color: var(--border-color);
}

.sidebar-menu li.active {
  font-weight: 600;
  border-left-color: var(--text-color);
}

/* Content Area - Fills remaining space */
.settings-content {
  flex: 1; /* Takes all remaining width */
  padding: 30px 40px;
  display: flex;
  flex-direction: column;
  
  /* SCROLLBAR FIX: 
     'overflow-y: scroll' forces the scrollbar track to always be visible.
     This prevents the content width from 'jumping' when switching 
     between long pages (with scrollbar) and short pages (without).
  */
  overflow-y: scroll; 
}

/* Wrapper inside content to constrain form width, but centered */
.content-wrapper {
  max-width: 700px;
  width: 100%;
}

/* Titles */
.section-title {
  font-size: 20px;
  font-weight: 400;
  margin-bottom: 24px;
}

.section-desc {
  color: #8e8e8e;
  font-size: 14px;
  margin-bottom: 20px;
}

/* Edit Profile Header */
.edit-profile-view {
  display: flex;
  flex-direction: column;
}

.profile-header {
  display: flex;
  align-items: center;
  margin-bottom: 30px;
  margin-left: 10px;
}

.profile-avatar-container {
  width: 38px;
  margin-right: 32px;
  display: flex;
  justify-content: flex-end;
}

.avatar-large {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  object-fit: cover;
}

.profile-actions {
  display: flex;
  flex-direction: column;
}

.username-display {
  font-size: 20px;
  font-weight: 400;
  line-height: 22px;
  margin-bottom: 2px;
}

.link-btn {
  background: none;
  border: none;
  color: #0095f6;
  font-weight: 600;
  font-size: 14px;
  padding: 0;
  cursor: pointer;
  text-align: left;
}

.hidden-input {
  display: none;
}

/* Forms (Label Left, Input Right) */
.ig-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: flex;
  align-items: flex-start;
  margin-bottom: 16px;
}

.form-row aside {
  width: 120px; 
  padding-right: 32px;
  text-align: right;
  padding-top: 6px; 
}

.form-row aside label {
  font-size: 16px;
  font-weight: 600;
}

.form-row .field {
  flex: 1;
  max-width: 400px;
}

.form-row input[type="text"],
.form-row textarea,
.custom-select {
  width: 100%;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  color: var(--text-color);
  padding: 0 10px;
  font-size: 16px;
  height: 32px;
}

.form-row textarea {
  height: auto;
  padding: 10px;
  resize: vertical;
}

.custom-select {
  height: 32px;
}

.help-text {
  font-size: 12px;
  color: #8e8e8e;
  margin-top: 10px;
  line-height: 16px;
}

.char-count {
  text-align: right;
  font-size: 12px;
  color: #8e8e8e;
  margin-top: 5px;
}

/* Standard Inputs for other tabs */
.ig-input {
  width: 100%;
  background: var(--input-bg, #262626);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: white;
  padding: 9px 12px;
  font-size: 14px;
}
.search-container {
  margin-bottom: 20px;
}

/* Buttons */
.btn-primary {
  background-color: #0095f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 5px 9px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
}
.btn-primary:hover {
  background-color: #1877f2;
}

.btn-secondary {
  background-color: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 4px;
  padding: 5px 9px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
}

.btn-danger-outline {
  background-color: transparent;
  border: 1px solid #ed4956;
  color: #ed4956;
  border-radius: 4px;
  padding: 5px 9px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
}

.btn-text-blue {
  background: none;
  border: none;
  color: #0095f6;
  font-weight: 600;
  cursor: pointer;
}

.submit-row .btn-primary {
  padding: 8px 24px; 
}

.full-width {
  width: 100%;
  padding: 10px;
  margin-top: 10px;
}

/* Lists */
.user-list {
  display: flex;
  flex-direction: column;
}

.user-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
}

.user-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar-small {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
}

.username {
  font-weight: 600;
  font-size: 14px;
}

.search-results {
  background-color: var(--dropdown-bg, #262626);
  border-radius: 6px;
  padding: 8px;
  margin-bottom: 20px;
}

.sub-header {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
  margin-top: 10px;
}

.empty-state {
  color: #8e8e8e;
  font-size: 14px;
  padding: 20px 0;
  text-align: center;
}

/* Settings Items */
.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.setting-info label {
  font-size: 16px;
  font-weight: 400;
  display: block;
  margin-bottom: 4px;
}

.setting-info p {
  font-size: 12px;
  color: #8e8e8e;
  max-width: 350px;
}

/* Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #363636; 
  transition: .4s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #0095f6;
}

input:checked + .slider:before {
  transform: translateX(16px);
}

/* Stack Form */
.stack-form .form-group {
  margin-bottom: 16px;
}
.stack-form label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  font-size: 14px;
}
.file-control {
  margin-top: 5px;
}

/* Responsive */
@media (max-width: 900px) {
  .settings-container {
    flex-direction: column;
    border: none;
    height: auto;
    width: 100%;
    min-width: 0; /* Remove restriction on mobile */
    max-width: 100%;
    margin: 0;
    overflow: visible;
  }
  .settings-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
    flex-direction: row;
    overflow-x: auto;
  }
  .sidebar-menu {
    display: flex;
  }
  .sidebar-menu li {
    padding: 15px;
    border-left: none;
    border-bottom: 2px solid transparent;
    white-space: nowrap;
  }
  .sidebar-menu li.active {
    border-bottom-color: var(--text-color);
    border-left: none;
  }
  .settings-content {
    padding: 20px;
    overflow-y: visible;
  }
  .form-row {
    flex-direction: column;
  }
  .form-row aside {
    width: 100%;
    text-align: left;
    margin-bottom: 8px;
    padding: 0;
  }
  .form-row .field {
    max-width: 100%;
  }
  .profile-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .profile-avatar-container {
    margin-right: 0;
    margin-bottom: 10px;
  }
}
</style>