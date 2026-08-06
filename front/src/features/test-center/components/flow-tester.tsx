'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Badge } from '@minisource/ui';
import { Play, Clock, ShieldCheck } from 'lucide-react';
import { FlowDefinition, FlowExecutionState, FlowStepResult } from '../types';
import { executeFlowStep } from '../lib/flow-runner';

interface FlowTesterProps {
  flows: FlowDefinition[];
  requestConfig: {
    authMode: any;
    customToken: string;
    headersMap: Record<string, string>;
    targetMode: 'direct' | 'gateway';
  };
  onRequestConfirmation: (flow: FlowDefinition) => void;
}

export function FlowTester({ flows, requestConfig, onRequestConfirmation }: FlowTesterProps) {
  const [selectedFlowId, setSelectedFlowId] = useState(flows[0]?.id || '');
  const [executionState, setExecutionState] = useState<FlowExecutionState | null>(null);
  const [running, setRunning] = useState(false);

  const selectedFlow = flows.find((f) => f.id === selectedFlowId) || flows[0];

  const runFlowSequence = async (flowToRun: FlowDefinition) => {
    setRunning(true);
    const startTime = performance.now();
    const initialState: FlowExecutionState = {
      flowId: flowToRun.id,
      state: 'running',
      currentStepIndex: 0,
      stepResults: flowToRun.steps.map((s) => ({
        stepId: s.id,
        title: s.title,
        state: 'idle',
      })),
      extractedContext: {},
      createdResourceIds: [],
      startTime: new Date().toLocaleTimeString(),
    };

    setExecutionState(initialState);
    let currentContext: Record<string, any> = {};
    const stepResults: FlowStepResult[] = [];
    let flowFailed = false;

    // Execute Main Flow Steps
    for (let i = 0; i < flowToRun.steps.length; i++) {
      const stepDef = flowToRun.steps[i];

      setExecutionState((prev) =>
        prev
          ? {
              ...prev,
              currentStepIndex: i,
              stepResults: prev.stepResults.map((r, idx) => (idx === i ? { ...r, state: 'running' } : r)),
            }
          : prev
      );

      const { stepResult, newExtracted } = await executeFlowStep(stepDef, currentContext, requestConfig);
      currentContext = { ...currentContext, ...newExtracted };
      stepResults.push(stepResult);

      setExecutionState((prev) =>
        prev
          ? {
              ...prev,
              stepResults: prev.stepResults.map((r, idx) => (idx === i ? stepResult : r)),
              extractedContext: currentContext,
            }
          : prev
      );

      if (stepResult.state === 'failed' && !stepDef.allowFailure) {
        flowFailed = true;
        break;
      }
    }

    // Execute Cleanup Steps if defined
    const cleanupResults: FlowStepResult[] = [];
    if (flowToRun.cleanupSteps && flowToRun.cleanupSteps.length > 0) {
      for (const cleanupDef of flowToRun.cleanupSteps) {
        const { stepResult } = await executeFlowStep(cleanupDef, currentContext, requestConfig);
        cleanupResults.push(stepResult);
      }
    }

    const totalDuration = Math.round(performance.now() - startTime);
    setExecutionState((prev) =>
      prev
        ? {
            ...prev,
            state: flowFailed ? 'failed' : 'passed',
            cleanupResults,
            endTime: new Date().toLocaleTimeString(),
            totalDuration,
          }
        : prev
    );

    setRunning(false);
  };

  const handleStartFlow = (flowToRun: FlowDefinition) => {
    if (flowToRun.requiresConfirmation || flowToRun.safetyClass === 'EXTERNAL DELIVERY') {
      onRequestConfirmation(flowToRun);
    } else {
      runFlowSequence(flowToRun);
    }
  };

  return (
    <div className="space-y-6">
      <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
        <CardHeader className="py-4 border-b border-border/50">
          <CardTitle className="text-sm font-semibold flex items-center justify-between">
            <span className="flex items-center gap-2">
              <Play className="h-4 w-4 text-primary" /> Multi-Step Business Flow Test Center
            </span>
            <Badge variant="outline">{flows.length} Flow Scenarios Available</Badge>
          </CardTitle>
          <CardDescription className="text-xs">
            Run end-to-end integration scenarios verifying multi-step state transitions, value extractions, and assertions.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4 pt-4">
          {/* Flow Selection Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {flows.map((flow) => (
              <div
                key={flow.id}
                onClick={() => setSelectedFlowId(flow.id)}
                className={`p-3.5 rounded-xl border cursor-pointer transition-all space-y-2 ${
                  selectedFlowId === flow.id
                    ? 'border-primary bg-primary/5 shadow-sm'
                    : 'border-border/60 hover:border-border bg-muted/10'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-xs text-foreground truncate max-w-[220px]">{flow.name}</span>
                  <Badge variant="outline" className="text-[10px] uppercase font-mono px-1.5 py-0">
                    {flow.steps.length} Steps
                  </Badge>
                </div>
                <p className="text-[11px] text-muted-foreground line-clamp-2 leading-relaxed">{flow.description}</p>
              </div>
            ))}
          </div>

          {selectedFlow && (
            <div className="p-4 rounded-xl border border-border/60 bg-muted/20 space-y-4">
              <div className="flex items-center justify-between border-b border-border/40 pb-3">
                <div>
                  <h4 className="font-bold text-sm text-foreground">{selectedFlow.name}</h4>
                  <p className="text-xs text-muted-foreground mt-0.5">{selectedFlow.description}</p>
                </div>
                <Button
                  onClick={() => handleStartFlow(selectedFlow)}
                  disabled={running}
                  size="sm"
                  className="gap-1.5 text-xs font-semibold bg-primary hover:bg-primary/90"
                >
                  <Play className="h-3.5 w-3.5" /> {running ? 'Executing Scenario...' : 'Run Scenario'}
                </Button>
              </div>

              {/* Visual Stepper List */}
              <div className="space-y-3">
                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block">
                  Scenario Execution Timeline
                </span>

                <div className="space-y-2">
                  {selectedFlow.steps.map((step, idx) => {
                    const stepRes = executionState?.stepResults.find((r) => r.stepId === step.id);
                    const state = stepRes?.state || 'idle';

                    return (
                      <div
                        key={step.id}
                        className={`p-3 rounded-lg border text-xs transition-colors ${
                          state === 'success'
                            ? 'bg-emerald-500/5 border-emerald-500/30'
                            : state === 'failed'
                            ? 'bg-rose-500/5 border-rose-500/30'
                            : state === 'running'
                            ? 'bg-primary/10 border-primary animate-pulse'
                            : 'bg-card border-border/40'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2.5">
                            <span
                              className={`h-5 w-5 rounded-full flex items-center justify-center text-[10px] font-bold ${
                                state === 'success'
                                  ? 'bg-emerald-500 text-white'
                                  : state === 'failed'
                                  ? 'bg-rose-500 text-white'
                                  : state === 'running'
                                  ? 'bg-primary text-white'
                                  : 'bg-muted text-muted-foreground'
                              }`}
                            >
                              {idx + 1}
                            </span>
                            <span className="font-semibold text-foreground">{step.title}</span>
                          </div>

                          <div className="flex items-center gap-2 font-mono text-[10px]">
                            {stepRes?.statusCode && <span>HTTP {stepRes.statusCode}</span>}
                            {stepRes?.duration && (
                              <span className="text-muted-foreground flex items-center gap-1">
                                <Clock className="h-3 w-3" /> {stepRes.duration}ms
                              </span>
                            )}
                            <Badge variant="outline" className="capitalize text-[10px]">
                              {state}
                            </Badge>
                          </div>
                        </div>

                        <p className="text-[11px] text-muted-foreground mt-1 pl-7">{step.description}</p>

                        {/* Step Failures & Evidence */}
                        {stepRes?.errorMessage && (
                          <div className="mt-2 ml-7 p-2 rounded bg-rose-500/10 text-rose-600 dark:text-rose-400 font-mono text-[11px]">
                            ❌ {stepRes.errorMessage}
                          </div>
                        )}

                        {/* Extracted Variables Badges */}
                        {stepRes?.extractedValues && Object.keys(stepRes.extractedValues).length > 0 && (
                          <div className="mt-2 ml-7 flex flex-wrap gap-1.5 items-center">
                            <span className="text-[10px] text-muted-foreground font-mono">Extracted:</span>
                            {Object.entries(stepRes.extractedValues).map(([k, v]) => (
                              <Badge key={k} variant="secondary" className="font-mono text-[10px] px-1.5 py-0">
                                {k} = {String(v)}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Cleanup Results if available */}
              {executionState?.cleanupResults && executionState.cleanupResults.length > 0 && (
                <div className="pt-3 border-t border-border/40 space-y-2">
                  <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-1">
                    <ShieldCheck className="h-3.5 w-3.5 text-emerald-500" /> Automated Resource Cleanup Results
                  </span>
                  <div className="space-y-1.5">
                    {executionState.cleanupResults.map((cRes) => (
                      <div key={cRes.stepId} className="p-2 rounded bg-muted/30 border border-border/40 text-xs flex justify-between font-mono">
                        <span>{cRes.title}</span>
                        <span className={cRes.state === 'success' ? 'text-emerald-500' : 'text-rose-500'}>
                          {cRes.state}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
