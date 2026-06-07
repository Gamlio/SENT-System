import React from 'react';
import { ShieldCheck, Monitor, ShieldAlert } from 'lucide-react';

export const riskScoringLevels = [
  { label: 'Rủi Ro Thấp', value: 'Low', color: '#10b981', icon: React.createElement(ShieldCheck) },
  { label: 'Rủi Ro Trung Bình', value: 'Medium', color: '#f59e0b', icon: React.createElement(Monitor) },
  { label: 'Rủi Ro Cao', value: 'High', color: '#ef4444', icon: React.createElement(ShieldAlert) },
  { label: 'Nguy Hiểm', value: 'Critical', color: '#be123c', icon: React.createElement(ShieldAlert) },
];

// Convert numeric risk score (0-100) to level metadata
export function getRiskLevel(score) {
  const s = Number(score) || 0;
  if (s <= 30) return riskScoringLevels[0];
  if (s > 30 && s <= 70) return riskScoringLevels[1];
  if (s > 70 && s <= 90) return riskScoringLevels[2];
  return riskScoringLevels[3];
}

export function getRiskLabel(score) {
  return getRiskLevel(score).value;
}

export default { riskScoringLevels, getRiskLevel, getRiskLabel };
