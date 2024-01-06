/*
 * Copyright (c) Joseph Prichard 2023
 */

import "./NotificationPanel.css";
import { Accessor, createEffect, createSignal, onCleanup } from "solid-js";

interface Props {
    errorMsg: Accessor<string>;
    onClose: () => void;
}

const NotificationPanel = ({ errorMsg, onClose }: Props) => {
    return (
        <div class="Notification" style={{display: errorMsg() == "" ? "none" : "block"}}>
            <div class="X" onClick={onClose}/>
            <div class="NotificationTitle">
                Error Occurred
            </div>
            <div>
                { errorMsg() }
            </div>
        </div>
    );
}

export interface TempMsg {
    msg: Accessor<string>;
    addMsg: (msg: string) => void;
    clearMsg: () => void;
}

export const useTempMsg = (time: number): TempMsg => {
    const [msg, setMsg] = createSignal<string>("");

    let timeout: ReturnType<typeof setTimeout> | undefined = undefined;

    const addMsg = (m: string) => {
        if (timeout) {
            clearTimeout(timeout);
        }
        timeout = setTimeout(clearMsg, time);
        setMsg(m);
    }

    const clearMsg = () => {
       addMsg("");
    }

    createEffect(() => {
        const _ = msg();
        const t = setTimeout(clearMsg, time);
        onCleanup(() => clearTimeout(t));
    });

    return {msg, addMsg, clearMsg};
}

export const defaultTempMsg = (): TempMsg => ({
    addMsg: (_: string) => {},
    msg: () => "",
    clearMsg: () => {}
})

export default NotificationPanel;