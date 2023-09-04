import { DOMAIN } from "../../App";

export async function createRoom(): Promise<[string, boolean]> {
    try {
        const resp = await fetch(`http://${DOMAIN}/rooms/create`, { method: "POST", mode: "cors" });
        const json = await resp.json();
        if (resp.ok) {
            return [json["Code"] as string, false];
        } else {
            return [json["ErrorDesc"] as string, true];
        }
    } catch(ex) {
        return ["Failed to create the room", true];
    }
}

export function setLocalPlayer(player: string) {
    localStorage.setItem("player", player);
}

export function getLocalPlayer() {
    return localStorage.getItem("player") ?? "";
}