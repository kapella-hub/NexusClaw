"use client";

import { useState, useEffect, useRef } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import { Wrench, Database, Send, Play, ChevronDown, ChevronRight, Activity, Terminal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { getToken } from "@/lib/api";

interface MCPInspectorProps {
  serverId: string;
}

export function MCPInspector({ serverId }: MCPInspectorProps) {
  const [url, setUrl] = useState("");
  
  useEffect(() => {
    // Determine the WS URL dynamically based on the current window location
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    // We connect to the backend API which proxies to the MCP server
    // For local dev with Next.js, we assume the backend is on port 8080 or use an env var
    const apiUrl = process.env.NEXT_PUBLIC_API_URL?.replace(/^http/, "ws") || `${protocol}//localhost:8080`;
    setUrl(`${apiUrl}/api/v1/nodes/${serverId}/ws?token=${getToken() || ""}`);
  }, [serverId]);

  const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket(
    url,
    {
      share: false,
      shouldReconnect: () => true,
      reconnectAttempts: 10,
      reconnectInterval: 3000,
      filter: () => false, // We handle messages manually via useEffect instead of driving state directly from lastJsonMessage if it gets too complex, but here we can just watch it.
    },
    url !== "" // Only connect when URL is ready
  );

  const [logs, setLogs] = useState<Array<{ type: 'send' | 'receive' | 'system'; data: any; timestamp: Date }>>([]);
  const [tools, setTools] = useState<any[]>([]);
  const scrollRef = useRef<HTMLDivElement>(null);
  const [requestId, setRequestId] = useState(1);
  const [activeTab, setActiveTab] = useState<'tools' | 'logs'>('tools');

  // Request tools when connection opens
  useEffect(() => {
    if (readyState === ReadyState.OPEN) {
      setLogs(prev => [...prev, { type: 'system', data: "Connected to MCP Server WebSocket", timestamp: new Date() }]);
      
      const req = {
        jsonrpc: "2.0",
        id: "init",
        method: "tools/list"
      };
      
      sendJsonMessage(req);
      setLogs(prev => [...prev, { type: 'send', data: req, timestamp: new Date() }]);
    } else if (readyState === ReadyState.CLOSED) {
      setLogs(prev => [...prev, { type: 'system', data: "Disconnected from MCP Server", timestamp: new Date() }]);
    }
  }, [readyState, sendJsonMessage]);

  // Handle incoming messages
  useEffect(() => {
    if (lastJsonMessage) {
      setLogs(prev => [...prev, { type: 'receive', data: lastJsonMessage, timestamp: new Date() }]);
      
      const msg = lastJsonMessage as any;
      if (msg.id === "init" && msg.result && msg.result.tools) {
        setTools(msg.result.tools);
      }
    }
  }, [lastJsonMessage]);

  // Auto-scroll logs
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs, activeTab]);

  const executeTool = (toolName: string, schema: any) => {
    const args: Record<string, any> = {};
    
    // Auto-fill mock data for required args for testing
    if (schema?.properties) {
      Object.keys(schema.properties).forEach(key => {
        const prop = schema.properties[key];
        if (prop.type === "string") args[key] = "test string";
        if (prop.type === "number") args[key] = 1;
        if (prop.type === "boolean") args[key] = true;
      });
    }

    const reqId = requestId.toString();
    setRequestId(prev => prev + 1);

    const req = {
      jsonrpc: "2.0",
      id: reqId,
      method: "tools/call",
      params: {
        name: toolName,
        arguments: args
      }
    };

    sendJsonMessage(req);
    setLogs(prev => [...prev, { type: 'send', data: req, timestamp: new Date() }]);
    setActiveTab('logs');
  };

  const connectionStatus = {
    [ReadyState.CONNECTING]: { label: 'Connecting...', color: 'text-yellow-500', bg: 'bg-yellow-500' },
    [ReadyState.OPEN]: { label: 'Connected', color: 'text-green-500', bg: 'bg-green-500' },
    [ReadyState.CLOSING]: { label: 'Closing...', color: 'text-orange-500', bg: 'bg-orange-500' },
    [ReadyState.CLOSED]: { label: 'Disconnected', color: 'text-red-500', bg: 'bg-red-500' },
    [ReadyState.UNINSTANTIATED]: { label: 'Uninstantiated', color: 'text-gray-500', bg: 'bg-gray-500' },
  }[readyState];

  return (
    <Card className="flex h-[600px] flex-col shadow-lg border-primary/20 bg-gradient-to-br from-background to-muted/20">
      <CardHeader className="border-b pb-4 flex flex-row items-center justify-between space-y-0">
        <div className="space-y-1">
          <CardTitle className="text-xl flex items-center gap-2">
            <Activity className="size-5 text-primary" />
            Interactive MCP Inspector
          </CardTitle>
          <CardDescription>
            Test and interact with your server directly via WebSocket JSON-RPC
          </CardDescription>
        </div>
        <div className="flex items-center gap-2 text-sm font-medium">
          <span className="relative flex h-3 w-3">
            {readyState === ReadyState.OPEN && <span className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 bg-green-400"></span>}
            <span className={`relative inline-flex rounded-full h-3 w-3 ${connectionStatus.bg}`}></span>
          </span>
          <span className={connectionStatus.color}>{connectionStatus.label}</span>
        </div>
      </CardHeader>
      
      <div className="flex border-b">
        <button 
          onClick={() => setActiveTab('tools')}
          className={`flex-1 py-3 text-sm font-medium transition-colors flex items-center justify-center gap-2 ${activeTab === 'tools' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'}`}
        >
          <Wrench className="size-4" />
          Available Tools ({tools.length})
        </button>
        <button 
          onClick={() => setActiveTab('logs')}
          className={`flex-1 py-3 text-sm font-medium transition-colors flex items-center justify-center gap-2 ${activeTab === 'logs' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'}`}
        >
          <Terminal className="size-4" />
          RPC Logs ({logs.length})
        </button>
      </div>

      <CardContent className="flex-1 overflow-hidden p-0">
        {activeTab === 'tools' ? (
          <ScrollArea className="h-full p-6">
            <div className="grid gap-4 sm:grid-cols-2">
              {tools.length === 0 ? (
                <div className="col-span-2 text-center py-12 text-muted-foreground">
                  {readyState === ReadyState.OPEN ? "No tools discovered (or awaiting response)..." : "Connect to discover tools."}
                </div>
              ) : (
                tools.map((tool, idx) => (
                  <div key={idx} className="rounded-lg border bg-card text-card-foreground shadow-sm p-4 hover:shadow-md transition-shadow group">
                    <div className="flex justify-between items-start mb-2">
                      <h4 className="font-semibold text-lg tracking-tight group-hover:text-primary transition-colors">{tool.name}</h4>
                      <Button variant="secondary" size="sm" onClick={() => executeTool(tool.name, tool.inputSchema)}>
                        <Play className="size-3 mr-1" />
                        Run
                      </Button>
                    </div>
                    <p className="text-sm text-muted-foreground mb-4 line-clamp-2" title={tool.description}>
                      {tool.description || "No description provided."}
                    </p>
                    <div className="flex flex-wrap gap-2">
                      {tool.inputSchema?.properties && Object.keys(tool.inputSchema.properties).map(prop => (
                        <Badge key={prop} variant="outline" className="text-xs bg-muted/50">
                          {prop} ({tool.inputSchema.properties[prop].type})
                        </Badge>
                      ))}
                    </div>
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        ) : (
          <div className="flex h-full flex-col p-4 bg-muted/10 font-mono text-sm">
            <div ref={scrollRef} className="flex-1 overflow-y-auto space-y-4 pr-4">
              {logs.map((log, idx) => (
                <div key={idx} className={`flex flex-col rounded-md p-3 border ${log.type === 'send' ? 'bg-primary/5 border-primary/20 ml-8' : log.type === 'receive' ? 'bg-card border-border mr-8' : 'bg-muted border-muted-foreground/20 items-center justify-center text-muted-foreground'}`}>
                  <div className="flex justify-between items-center mb-1 text-xs opacity-60">
                    <span className="uppercase font-bold tracking-wider">
                      {log.type === 'send' ? 'Request' : log.type === 'receive' ? 'Response' : 'System'}
                    </span>
                    <span>{log.timestamp.toLocaleTimeString()}</span>
                  </div>
                  {log.type !== 'system' ? (
                    <pre className="whitespace-pre-wrap break-all overflow-x-auto text-[13px]">
                      {JSON.stringify(log.data, null, 2)}
                    </pre>
                  ) : (
                    <span className="text-center">{log.data}</span>
                  )}
                </div>
              ))}
              {logs.length === 0 && (
                <div className="h-full flex items-center justify-center text-muted-foreground opacity-50">
                  No RPC traffic yet.
                </div>
              )}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
