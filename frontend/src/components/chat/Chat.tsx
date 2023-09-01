import "./Chat.css"
import { For, createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import { RoomProps } from "../../pages/room/Room";
import { CHAT_CODE } from "../../pages/room/messages";

export interface ChatMsg {
    text: string;
    player: string;
    guessIncScore: number;
}

const Chat = ({ room }: RoomProps) => {
    const [newChat, setNewChat] = createSignal("");
    const [chats, setChats] = createStore<ChatMsg[]>([]);

    room.subscribe(CHAT_CODE, (payload) => {

    });

    return (
        <div class="Chat">
            <div>
                <For each={chats}>
                    {(chat) => <ChatMsg {...chat}/>}
                </For>
            </div>
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
            <span class="BoldText">
                {chat.player}:
            </span>
            <span>
                {" "}
            </span>
            <span>
                {chat.text}
            </span>
        </div>
    );
}

export default Chat;