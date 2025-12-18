import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import PostComponent from "../PostComponent.vue";
import { postsApi } from "../../services/apiService";

vi.mock("vue-router", () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

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

vi.mock("../../composables/useAuth", () => ({
  useAuth: () => ({
    user: { value: { id: "user-123" } },
  }),
}));

vi.stubGlobal(
  "confirm",
  vi.fn(() => true)
);
vi.stubGlobal("alert", vi.fn());

describe("PostComponent", () => {
  const mockPost = {
    id: "post-abc",
    user_id: "user-123", 
    username: "testuser",
    caption: "Hello World",
    created_at: new Date().toISOString(),
    likes_count: 10,
    comments_count: 2,
    media: [{ media_url: "http://img.com/1.jpg", media_type: "image/jpeg" }],
  };

  const setAuthToken = (userId: string) => {
    const payload = JSON.stringify({ user_id: userId });
    const token = `header.${window.btoa(payload)}.signature`;
    localStorage.setItem("accessToken", token);
  };

  beforeEach(() => {
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

    await wrapper.find(".more-button").trigger("click");

    const deleteBtn = wrapper.find(".menu-item.delete");
    expect(deleteBtn.exists()).toBe(true);
    expect(deleteBtn.text()).toBe("Remove Post");
  });

  it('does NOT render "Remove Post" when user is NOT owner', async () => {
    setAuthToken("user-999");

    const wrapper = mount(PostComponent, {
      props: { post: mockPost },
    });

    await wrapper.find(".more-button").trigger("click");

    const deleteBtn = wrapper.find(".menu-item.delete");
    expect(deleteBtn.exists()).toBe(false);

    const reportBtn = wrapper.find(".menu-item.report");
    expect(reportBtn.exists()).toBe(true);
  });

  it('emits "post-deleted" event when delete is confirmed', async () => {
    const wrapper = mount(PostComponent, {
      props: { post: mockPost },
    });

    await wrapper.find(".more-button").trigger("click");

    await wrapper.find(".menu-item.delete").trigger("click");

    expect(postsApi.deletePost).toHaveBeenCalledWith("post-abc");

    expect(wrapper.emitted()).toHaveProperty("post-deleted");
    expect(wrapper.emitted("post-deleted")![0]).toEqual(["post-abc"]);
  });
});
