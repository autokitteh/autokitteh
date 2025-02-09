import axios from "axios";

export async function getData(args: any): Promise<string> {
    console.log(args);
    return await axios.get('https://jsonplaceholder.typicode.com/posts/1');
}

