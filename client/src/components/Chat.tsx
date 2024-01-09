/*
 * Copyright (c) Joseph Prichard 2023
 */

import "./Chat.css"
import { For, createSignal, Show } from "solid-js";
import { RoomProps } from "../pages/Room";
import { CHAT_CODE, Player, STATE_CODE, TEXT_CODE } from "../websocket/messages";
import { StateMsg } from "../websocket/state";

export interface ChatMsg {
    text: string;
    player: Player;
    guessPointsInc: number;
}

const Chat = ({ room }: RoomProps) => {
    const [newChat, setNewChat] = createSignal("");
    const [chats, setChats] = createSignal<ChatMsg[]>([]);

    room.subscribe<StateMsg>(STATE_CODE, (payload) => {
        const msg = payload.msg;
        setChats(msg.chatLog);
    });

    room.subscribe<ChatMsg>(CHAT_CODE, (payload) => {
        const msg = payload.msg;
        setChats([...chats(), msg]);
    });

    return (
        <div class="Chat">
            <div class="TopTitle">
                Chat
            </div>
            <div>
                <For each={chats()}>
                    {(chat) => <ChatMsg {...chat}/>}
                </For>
            </div>
            <input
                placeholder="Type your guess here"
                class="ChatInput"
                value={newChat()}
                onInput={e => setNewChat(e.target.value)}
                onKeyDown={e => {
                    if (e.key === "Enter") {
                        room.send({
                            code: TEXT_CODE,
                            msg: {
                                text: newChat()
                            }
                        });
                    }
                }}
            />
        </div>
    );
}

const ChatMsg = (props: ChatMsg) => {
    return (
        <Show
            when={props.guessPointsInc === 0}
            fallback={
                <div class="ChatMsg">
                    <span>
                        {props.player.name} guessed the word! (+{props.guessPointsInc})
                    </span>
                </div>
            }
        >
            <div class="ChatMsg">
                <span>
                    {props.player.name}:
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