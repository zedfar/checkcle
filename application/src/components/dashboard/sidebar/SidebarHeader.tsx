
import React from "react";
import { useTheme } from "@/contexts/ThemeContext";

interface SidebarHeaderProps {
  collapsed: boolean;
}

export const SidebarHeader: React.FC<SidebarHeaderProps> = ({ collapsed }) => {
  const { theme } = useTheme();

  return (
    <div className={`p-4 ${theme === 'dark' ? 'border-[#1e1e1e]' : 'border-sidebar-border'} border-b flex items-center ${collapsed ? 'justify-center' : ''}`}>
      <div className="h-8 w-8 bg-gray-600 rounded flex items-center justify-center mr-2">
        <img
          src="/favicon.svg"
          alt="CheckCle MOD"
          className="h-6 w-6"
        />
      </div>
      {!collapsed && (
        <div className="flex flex-col">
          <h1 className="text-xl font-semibold leading-tight">CheckCle App</h1>
          <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 w-fit mt-0.5">MOD</span>
        </div>
      )}
    </div>
  );
};