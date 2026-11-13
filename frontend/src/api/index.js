// Server -> client message types. Must match backend/pkg/websocket/protocol.go.
export const MsgType = {
    CHAT: 1,
    GAME_EVENT: 2,
    ROOM_CREATED: 3,
    ERROR: 4,
    GAME_STATE: 5,
    ROOM_INFO: 6,
    WELCOME: 7,
    LOGIN_FAILED: 8,
};

const TOKEN_KEY = "token";
const socket = new WebSocket("ws://localhost:8080/ws");
let greeted = false; // StrictMode calls connect() twice; only greet the server once

function onOpen(cb) {
    if (greeted) return;
    greeted = true;

    const token = localStorage.getItem(TOKEN_KEY);
    if (token) {
        send({ action: "RESUME", token });
    } else {
        cb({ type: MsgType.LOGIN_FAILED }); // no token yet: ask for a name
    }
}

export function connect(cb) {
    socket.onopen = () => onOpen(cb);

    socket.onmessage = event => {
        const msg = JSON.parse(event.data);
        if (msg.type === MsgType.WELCOME) localStorage.setItem(TOKEN_KEY, msg.body.token);
        if (msg.type === MsgType.LOGIN_FAILED) localStorage.removeItem(TOKEN_KEY);
        cb(msg);
    };

    socket.onclose = event => console.log("Socket closed:", event);
    socket.onerror = error => console.log("Socket error:", error);

    // If the socket opened before connect() was called, onopen never fires.
    if (socket.readyState === WebSocket.OPEN) onOpen(cb);
}

export function send(payload) {
    if (socket.readyState !== WebSocket.OPEN) {
        console.warn("Socket not open, dropping message:", payload);
        return;
    }
    socket.send(JSON.stringify(payload));
}

export function login(name) {
    send({ action: "LOGIN", name });
}

export function logout() {
    localStorage.removeItem(TOKEN_KEY);
    window.location.reload();
}
