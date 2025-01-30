import axios from "axios";

export async function getData(): Promise<string> {
    return await axios.get('https://jsonplaceholder.typicode.com/posts/1');
}

