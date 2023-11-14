import "./Chat.css"
import { For, createSignal, Show } from "solid-js";
import { createStore } from "solid-js/store";
import { RoomProps } from "../pages/Room";
import { CHAT_CODE } from "../pages/messages";

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
            <div class="TopTitle">
                Chat
            </div>
            <div>
                <For each={chats}>
                    {(chat) => (
                        <ChatMsg {...chat}/>
                    )}
                </For>
            </div>
            <input
                placeholder="Type your guess here"
                class="ChatInput"
                value={newChat()}
                onChange={e => setNewChat(e.target.value)}
            />
        </div>
    );
}

const ChatMsg = (props: ChatMsg) => {
    return (
        <Show
            when={props.guessIncScore <= 0}
            fallback={
                <div class="ChatMsg">
                    <span>
                        {props.player} guessed the word! (+{props.guessIncScore})
                    </span>
                </div>
            }
        >
            <div class="ChatMsg">
                <span>
                    {props.player}:
                </span>
                <span>
                    {" "}
                </span>
                <span>
                    {props.text}
                </span>
            </div>
        </Show>
    );
}

export default Chat;