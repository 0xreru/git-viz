import { useCallback, useMemo } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type NodeTypes,
  MarkerType,
} from '@xyflow/react';
import dagre from 'dagre';
import '@xyflow/react/dist/style.css';

import CommitNode, { type CommitNodeData } from './CommitNode';
import type { ReconData, CommitNode as CommitNodeType } from '../types';

// Register custom node types
const nodeTypes: NodeTypes = {
  commit: CommitNode,
} as const;

// Dagre layout configuration
const NODE_WIDTH = 180;
const NODE_HEIGHT = 100;

function getLayoutedElements(
  nodes: Node[],
  edges: Edge[],
  direction: 'TB' | 'BT' = 'TB'
): { nodes: Node[]; edges: Edge[] } {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  
  dagreGraph.setGraph({ 
    rankdir: direction,
    nodesep: 50,
    ranksep: 80,
    marginx: 20,
    marginy: 20,
  });

  // Add nodes to dagre
  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  });

  // Add edges to dagre
  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  // Run layout algorithm
  dagre.layout(dagreGraph);

  // Apply calculated positions to nodes
  const layoutedNodes = nodes.map((node): Node => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - NODE_WIDTH / 2,
        y: nodeWithPosition.y - NODE_HEIGHT / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
}

interface GraphViewProps {
  data: ReconData;
  onNodeClick?: (commit: CommitNodeType) => void;
}

export default function GraphView({ data, onNodeClick }: GraphViewProps) {
  // Build hash -> commit lookup for quick access
  const commitMap = useMemo(() => {
    const map = new Map<string, CommitNodeType>();
    data.commits.forEach((c) => map.set(c.hash, c));
    data.orphans.forEach((c) => map.set(c.hash, c));
    return map;
  }, [data.commits, data.orphans]);

  // Create nodes from commits + orphans
  const initialNodes = useMemo((): Node[] => {
    const allCommits = [...data.commits, ...data.orphans];
    
    return allCommits.map((commit): Node => ({
      id: commit.hash,
      type: 'commit',
      position: { x: 0, y: 0 }, // Will be set by dagre
      data: { ...commit } as CommitNodeData,
    }));
  }, [data.commits, data.orphans]);

  // Create edges from parent relationships
  // Edge direction: parent -> child (source = parent, target = child)
  const initialEdges = useMemo((): Edge[] => {
    const allCommits = [...data.commits, ...data.orphans];
    const edges: Edge[] = [];
    const existingHashes = new Set(allCommits.map((c) => c.hash));

    allCommits.forEach((commit) => {
      commit.parents.forEach((parentHash) => {
        // Only create edge if parent exists in our dataset
        if (existingHashes.has(parentHash)) {
          edges.push({
            id: `${parentHash}->${commit.hash}`,
            source: parentHash,
            target: commit.hash,
            type: 'smoothstep',
            animated: commit.isOrphan,
            style: {
              stroke: commit.isOrphan ? '#f85149' : '#30363d',
              strokeWidth: commit.isOrphan ? 2 : 1.5,
            },
            markerEnd: {
              type: MarkerType.ArrowClosed,
              color: commit.isOrphan ? '#f85149' : '#484f58',
              width: 15,
              height: 15,
            },
          });
        }
      });
    });

    return edges;
  }, [data.commits, data.orphans]);

  // Apply dagre layout
  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(
    () => getLayoutedElements(initialNodes, initialEdges, 'TB'),
    [initialNodes, initialEdges]
  );

  const [nodes, , onNodesChange] = useNodesState(layoutedNodes);
  const [edges, , onEdgesChange] = useEdgesState(layoutedEdges);

  // Handle node click
  const handleNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      const commit = commitMap.get(node.id);
      if (commit && onNodeClick) {
        onNodeClick(commit);
      }
    },
    [commitMap, onNodeClick]
  );

  return (
    <div style={{ width: '100%', height: '100%', background: '#0d1117' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        nodeTypes={nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.1}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'smoothstep',
        }}
        proOptions={{ hideAttribution: true }}
      >
        <Background color="#30363d" gap={20} size={1} />
        <Controls 
          style={{ 
            background: '#161b22', 
            border: '1px solid #30363d',
            borderRadius: '6px',
          }}
        />
        <MiniMap
          style={{
            background: '#161b22',
            border: '1px solid #30363d',
            borderRadius: '6px',
          }}
          nodeColor={(node) => {
            const nodeData = node.data as CommitNodeData;
            if (nodeData?.isOrphan) return '#f85149';
            if (nodeData?.secretHits > 0) return '#d29922';
            if (nodeData?.isMerge) return '#a371f7';
            return '#30363d';
          }}
          maskColor="rgba(13, 17, 23, 0.8)"
        />
      </ReactFlow>
    </div>
  );
}
