<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { adminApi } from '@/services/apiService';

const activeTab = ref('users');
const reportType = ref<'post' | 'user'>('post');

// Data State
const users = ref<any[]>([]);
const verificationRequests = ref<any[]>([]);
const reports = ref<any[]>([]);
const newsletter = ref({ subject: '', body: '' });

// Fetch Data functions
const loadUsers = async () => {
  const res = await adminApi.getAllUsers();
  users.value = res.data;
};

const loadVerifications = async () => {
  const res = await adminApi.getVerificationRequests();
  verificationRequests.value = res.data;
};

const loadReports = async () => {
  const res = await adminApi.getReports(reportType.value);
  reports.value = res.data;
};

// Actions
const toggleBan = async (user: any, ban: boolean) => {
  if(!confirm(`Are you sure you want to ${ban ? 'ban' : 'unban'} this user?`)) return;
  await adminApi.banUser(user.user_id, ban);
  user.is_banned = ban;
};

const handleVerification = async (reqId: string, action: 'ACCEPTED' | 'REJECTED') => {
  await adminApi.reviewVerification(reqId, action);
  verificationRequests.value = verificationRequests.value.filter(r => r.id !== reqId);
  alert(`Request ${action.toLowerCase()} successfully`);
};

const handleReport = async (reportId: string, action: 'ACCEPT' | 'REJECT') => {
  await adminApi.reviewReport(reportId, reportType.value, action === 'ACCEPT' ? (reportType.value === 'post' ? 'DELETE_POST' : 'BAN_USER') : 'IGNORE');
  loadReports(); // Reload list
};

const sendNewsletter = async () => {
  await adminApi.sendNewsletter(newsletter.value.subject, newsletter.value.body);
  alert('Newsletter sent successfully!');
  newsletter.value = { subject: '', body: '' };
};

// Watchers and Mount
watch(activeTab, (newTab) => {
  if (newTab === 'users') loadUsers();
  if (newTab === 'verification') loadVerifications();
  if (newTab === 'reports') loadReports();
});

watch(reportType, () => {
  if (activeTab.value === 'reports') loadReports();
});

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="admin-container">
    <aside class="admin-sidebar">
      <h2>Admin Panel</h2>
      <ul>
        <li :class="{ active: activeTab === 'users' }" @click="activeTab = 'users'">User Management</li>
        <li :class="{ active: activeTab === 'verification' }" @click="activeTab = 'verification'">Verification Requests</li>
        <li :class="{ active: activeTab === 'reports' }" @click="activeTab = 'reports'">Content Moderation</li>
        <li :class="{ active: activeTab === 'newsletter' }" @click="activeTab = 'newsletter'">Newsletter</li>
      </ul>
    </aside>

    <main class="admin-content">
      <div v-if="activeTab === 'users'" class="tab-content">
        <h3>User Management</h3>
        <table class="data-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>Email</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.user_id">
              <td>{{ user.username }}</td>
              <td>{{ user.email }}</td>
              <td>
                <span :class="['status-badge', user.is_banned ? 'banned' : 'active']">
                  {{ user.is_banned ? 'Banned' : 'Active' }}
                </span>
              </td>
              <td>
                <button v-if="!user.is_banned" class="btn-danger" @click="toggleBan(user, true)">Ban</button>
                <button v-else class="btn-success" @click="toggleBan(user, false)">Unban</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="activeTab === 'verification'" class="tab-content">
        <h3>Verification Requests</h3>
        <div class="cards-grid">
          <div v-for="req in verificationRequests" :key="req.id" class="request-card">
            <div class="req-header">
              <img :src="req.profile_picture_url" alt="User" class="avatar">
              <strong>{{ req.username }}</strong>
            </div>
            <p><strong>Reason:</strong> {{ req.reason }}</p>
            <div class="id-photo-wrapper">
              <img :src="req.selfie_url" alt="ID Selfie" class="id-photo">
            </div>
            <div class="card-actions">
              <button class="btn-success" @click="handleVerification(req.id, 'ACCEPTED')">Accept</button>
              <button class="btn-danger" @click="handleVerification(req.id, 'REJECTED')">Reject</button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'reports'" class="tab-content">
        <div class="sub-tabs">
          <button :class="{ active: reportType === 'post' }" @click="reportType = 'post'">Post Reports</button>
          <button :class="{ active: reportType === 'user' }" @click="reportType = 'user'">User Reports</button>
        </div>

        <div class="list-container">
          <div v-for="report in reports" :key="report.id" class="report-item">
            <div class="report-info">
              <p><strong>Reporter:</strong> {{ report.reporter_name }}</p>
              <p><strong>Reason:</strong> {{ report.reason }}</p>
              <p v-if="reportType === 'post'"><strong>Post ID:</strong> {{ report.post_id }}</p>
              <p v-if="reportType === 'user'"><strong>Reported User:</strong> {{ report.reported_user_name }}</p>
            </div>
            <div class="report-actions">
              <button class="btn-danger" @click="handleReport(report.id, 'ACCEPT')">
                {{ reportType === 'post' ? 'Delete Post' : 'Ban User' }}
              </button>
              <button class="btn-secondary" @click="handleReport(report.id, 'REJECT')">Ignore</button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'newsletter'" class="tab-content">
        <h3>Send Newsletter</h3>
        <form class="newsletter-form" @submit.prevent="sendNewsletter">
          <div class="form-group">
            <label>Subject</label>
            <input v-model="newsletter.subject" type="text" required placeholder="Newsletter Subject">
          </div>
          <div class="form-group">
            <label>Message Body</label>
            <textarea v-model="newsletter.body" rows="6" required placeholder="Write your update here..."></textarea>
          </div>
          <button type="submit" class="btn-primary">Send to All Subscribers</button>
        </form>
      </div>
    </main>
  </div>
</template>

<style scoped>
.admin-container {
  display: flex;
  height: 100vh;
  background-color: #f5f5f5;
  color: #333;
}

.admin-sidebar {
  width: 250px;
  background: white;
  padding: 20px;
  border-right: 1px solid #dbdbdb;
}

.admin-sidebar ul {
  list-style: none;
  padding: 0;
}

.admin-sidebar li {
  padding: 15px;
  cursor: pointer;
  border-radius: 8px;
  margin-bottom: 5px;
}

.admin-sidebar li.active, .admin-sidebar li:hover {
  background-color: #efefef;
  font-weight: bold;
}

.admin-content {
  flex: 1;
  padding: 30px;
  overflow-y: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
}

.data-table th, .data-table td {
  padding: 12px 15px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.85rem;
}

.status-badge.active { background: #e0f8e0; color: #2e7d32; }
.status-badge.banned { background: #ffebee; color: #c62828; }

.cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.request-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.id-photo {
  width: 100%;
  height: 150px;
  object-fit: cover;
  border-radius: 4px;
  margin: 10px 0;
}

.report-item {
  background: white;
  padding: 15px;
  margin-bottom: 10px;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.btn-danger { background: #ff4444; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
.btn-success { background: #00C300; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
.btn-primary { background: #0095f6; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
.btn-secondary { background: #dbdbdb; color: #333; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }

.sub-tabs button {
  padding: 10px 20px;
  margin-right: 10px;
  border: none;
  background: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
}
.sub-tabs button.active {
  border-bottom-color: #000;
  font-weight: bold;
}
</style>