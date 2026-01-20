import axios from 'axios';


class UserService {
  private baseUrl: string;

  constructor(baseUrl = "https://jsonplaceholder.typicode.com") {
    this.baseUrl = baseUrl;
  }

  async getUserById(userId: number) {
    const response = await axios.get(`${this.baseUrl}/users/${userId}`);
    return {
      name: (response.data as any).name,
      email: (response.data as any).email,
      company: (response.data as any).company.name,
    };
  }

  async getUserPosts(userId: number) {
    const response = await axios.get(`${this.baseUrl}/users/${userId}/posts`);
    return (response.data as any[]).map((post: any) => ({
      title: post.title,
      body: post.body,
    }));
  }

  static async getAllUsers() {
    const response = await axios.get(
      "https://jsonplaceholder.typicode.com/users"
    );
    return (response.data as any[]).map((user: any) => user.name);
  }
}

export default UserService;
