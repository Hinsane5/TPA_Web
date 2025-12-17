import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PostComponent from "../PostComponent.vue";
import { postsApi } from "../../services/apiService";

// 1. Mock Vue Router to fix the injection warning
vi.mock("vue-router", () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

// 2. Mock API services
vi.mock("../../services/apiService", () => ({
  postsApi: {
    deletePost: vi.fn(() => Promise.resolve()),
    toggleSavePost: vi.fn(),
    getUserCollections: vi.fn(() => Promise.resolve({ data: [] })),
  },
  usersApi: {
    searchUsers: vi.fn(),
  },
  aiApi: {
    summarizeText: vi.fn(),
  },
}));

// 3. Mock useAuth (although isOwnPost uses localStorage, other parts might use this)
vi.mock("../../composables/useAuth", () => ({
  useAuth: () => ({
    user: { value: { id: "user-123" } },
  }),
}));

// 4. Stub Globals for jsdom
vi.stubGlobal(
  "confirm",
  vi.fn(() => true)
);
vi.stubGlobal("alert", vi.fn());

describe("PostComponent", () => {
  const mockPost = {
    id: "post-abc",
    user_id: "user-123", // Owner
    username: "testuser",
    caption: "Hello World",
    created_at: new Date().toISOString(),
    likes_count: 10,
    comments_count: 2,
    media: [{ media_url: "http://img.com/1.jpg", media_type: "image/jpeg" }],
  };

  // Helper to set a fake JWT in localStorage so getUserIdFromToken() works
  const setAuthToken = (userId: string) => {
    const payload = JSON.stringify({ user_id: userId });
    // Create a fake token structure: header.payload.signature
    // We use window.btoa to encode the payload to Base64
    const token = `header.${window.btoa(payload)}.signature`;
    localStorage.setItem("accessToken", token);
  };

  beforeEach(() => {
    // Default to being the owner
    setAuthToken("user-123");
  });

  afterEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('renders "Remove Post" option when user is owner', async () => {
    const wrapper = mount(PostComponent, {
      props: { post: mockPost },
    });

    // Open the menu
    await wrapper.find(".more-button").trigger("click");

    // Check if delete button exists
    const deleteBtn = wrapper.find(".menu-item.delete");
    expect(deleteBtn.exists()).toBe(true);
    expect(deleteBtn.text()).toBe("Remove Post");
  });

  it('does NOT render "Remove Post" when user is NOT owner', async () => {
    // Set token to a different user
    setAuthToken("user-999");

    const wrapper = mount(PostComponent, {
      props: { post: mockPost }, // Post belongs to 'user-123'
    });

    await wrapper.find(".more-button").trigger("click");

    const deleteBtn = wrapper.find(".menu-item.delete");
    expect(deleteBtn.exists()).toBe(false);

    // Should show Report button instead
    const reportBtn = wrapper.find(".menu-item.report");
    expect(reportBtn.exists()).toBe(true);
  });

  it('emits "post-deleted" event when delete is confirmed', async () => {
    const wrapper = mount(PostComponent, {
      props: { post: mockPost },
    });

    // 1. Open Menu
    await wrapper.find(".more-button").trigger("click");

    // 2. Click Delete
    await wrapper.find(".menu-item.delete").trigger("click");

    // 3. Verify API was called
    expect(postsApi.deletePost).toHaveBeenCalledWith("post-abc");

    // 4. Verify Event Emitted
    expect(wrapper.emitted()).toHaveProperty("post-deleted");
    expect(wrapper.emitted("post-deleted")![0]).toEqual(["post-abc"]);
  });
});
