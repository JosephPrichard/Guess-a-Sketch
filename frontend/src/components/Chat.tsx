import "./Chat.css"
import { For, createSignal } from "solid-js";
import { createStore } from "solid-js/store";

export interface ChatMsg {
    text: string;
    player: string;
    guessIncScore: number;
}

const Chat = () => {
    const [newChat, setNewChat] = createSignal("");
    const [chats, setChats] = createStore<ChatMsg[]>([
        { text: "123", player: "123", guessIncScore: 0},
        { text: "456", player: "123", guessIncScore: 0},
        { text: "789", player: "123", guessIncScore: 0}
    ]);

    return (
        <div class="Chat">
            <For each={chats}>
                {(chat) => <ChatMsg {...chat}/>}
            </For>
            <div class="Space"/>
            <input
                placeholder="Type your guess here"
                class="Input ChatInput"
                value={newChat()}
                onChange={e => setNewChat(e.target.value)}
            />
        </div>
    );
}

const ChatMsg = (chat: ChatMsg) => {
    return (
        <div class="ChatMsg">
            <span class="ChatPlayer">
                {chat.player}:
            </span>
            <span>
                {chat.text}
            </span>
        </div>
    );
}

export default Chat;