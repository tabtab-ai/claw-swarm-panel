import { Terminal as TerminalIcon } from "lucide-react";
import { ArrowLeft } from "lucide-react";
import { useEffect, useRef } from "react";
import { useNavigate, useParams } from "react-router";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";

import { getToken } from "../../utils/auth/token";
import { Button } from "./ui/button";

export function PodTerminal() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!id || !containerRef.current) return;

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: '"Cascadia Code", "Fira Code", monospace',
      theme: {
        background: "#0d1117",
        foreground: "#e6edf3",
        cursor: "#58a6ff",
        selectionBackground: "#264f78",
        black: "#484f58",
        red: "#ff7b72",
        green: "#3fb950",
        yellow: "#d29922",
        blue: "#58a6ff",
        magenta: "#bc8cff",
        cyan: "#39c5cf",
        white: "#b1bac4",
        brightBlack: "#6e7681",
        brightRed: "#ffa198",
        brightGreen: "#56d364",
        brightYellow: "#e3b341",
        brightBlue: "#79c0ff",
        brightMagenta: "#d2a8ff",
        brightCyan: "#56d4dd",
        brightWhite: "#f0f6fc",
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(containerRef.current);
    fitAddon.fit();

    // Build WebSocket URL — pass JWT as query param since browsers
    // cannot set custom headers on WebSocket connections.
    const token = getToken() ?? "";
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(
      `${proto}//${window.location.host}/claw/terminal/${id}?token=${encodeURIComponent(token)}`
    );
    ws.binaryType = "arraybuffer";

    ws.onopen = () => {
      const { cols, rows } = term;
      ws.send(JSON.stringify({ type: "resize", cols, rows }));
    };

    ws.onmessage = (e) => {
      const data =
        e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : e.data;
      term.write(data as Uint8Array);
    };

    ws.onerror = () => term.write("\r\n\x1b[31m[connection error]\x1b[0m\r\n");
    ws.onclose = () => term.write("\r\n\x1b[33m[disconnected]\x1b[0m\r\n");

    // stdin: send as binary
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(new TextEncoder().encode(data));
      }
    });

    // Resize: use xterm's built-in resize event for accuracy
    term.onResize(({ cols, rows }) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "resize", cols, rows }));
      }
    });

    const observer = new ResizeObserver(() => fitAddon.fit());
    observer.observe(containerRef.current!);

    return () => {
      ws.close();
      term.dispose();
      observer.disconnect();
    };
  }, [id]);

  return (
    <div className="flex flex-col h-full space-y-4">
      <div className="flex items-center gap-3">
        <Button
          variant="ghost"
          size="sm"
          className="gap-2"
          onClick={() => navigate(`/instances/${id}`)}
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <div className="flex items-center gap-2 text-slate-300">
          <TerminalIcon className="h-4 w-4 text-cyan-400" />
          <span className="font-mono text-sm">
            {id} <span className="text-slate-500">/ shell</span>
          </span>
        </div>
      </div>

      <div
        ref={containerRef}
        className="flex-1 rounded-lg border border-slate-700 overflow-hidden"
        style={{ minHeight: "500px", background: "#0d1117" }}
      />
    </div>
  );
}
