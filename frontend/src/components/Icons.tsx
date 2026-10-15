import React from 'react';

type P = { className?: string };

const make = (paths: React.ReactNode) => {
  const Icon: React.FC<P> = ({ className }) => (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className ?? 'h-5 w-5'}
      aria-hidden="true"
    >
      {paths}
    </svg>
  );
  return Icon;
};

export const IconDashboard = make(<><rect x="3" y="3" width="7" height="9" rx="2" /><rect x="14" y="3" width="7" height="5" rx="2" /><rect x="14" y="12" width="7" height="9" rx="2" /><rect x="3" y="16" width="7" height="5" rx="2" /></>);
export const IconBox = make(<><path d="M21 8l-9-5-9 5 9 5 9-5z" /><path d="M3 8v8l9 5 9-5V8" /><path d="M12 13v8" /></>);
export const IconWrench = make(<path d="M14.7 6.3a4 4 0 0 0-5.4 5.4L3 18l3 3 6.3-6.3a4 4 0 0 0 5.4-5.4l-2.6 2.6-2.4-.6-.6-2.4z" />);
export const IconLayers = make(<><path d="M12 3l9 5-9 5-9-5 9-5z" /><path d="M3 13l9 5 9-5" /></>);
export const IconUsers = make(<><circle cx="9" cy="8" r="3.5" /><path d="M2.5 20c.6-3.4 3.2-5.5 6.5-5.5s5.9 2.1 6.5 5.5" /><path d="M16 4.6a3.5 3.5 0 0 1 0 6.8" /><path d="M18 14.8c1.8.7 3 2.4 3.5 5.2" /></>);
export const IconShield = make(<><path d="M12 3l8 3v6c0 4.5-3.2 8-8 9-4.8-1-8-4.5-8-9V6l8-3z" /><path d="M9 12l2 2 4-4" /></>);
export const IconLogout = make(<><path d="M9 4H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h4" /><path d="M16 17l5-5-5-5" /><path d="M21 12H9" /></>);
export const IconMenu = make(<><path d="M4 7h16" /><path d="M4 12h16" /><path d="M4 17h10" /></>);
export const IconGrid = make(<><circle cx="6" cy="6" r="1.6" /><circle cx="12" cy="6" r="1.6" /><circle cx="18" cy="6" r="1.6" /><circle cx="6" cy="12" r="1.6" /><circle cx="12" cy="12" r="1.6" /><circle cx="18" cy="12" r="1.6" /><circle cx="6" cy="18" r="1.6" /><circle cx="12" cy="18" r="1.6" /><circle cx="18" cy="18" r="1.6" /></>);
export const IconX = make(<><path d="M6 6l12 12" /><path d="M18 6L6 18" /></>);
export const IconPlus = make(<><path d="M12 5v14" /><path d="M5 12h14" /></>);
export const IconSearch = make(<><circle cx="11" cy="11" r="7" /><path d="M20 20l-3.5-3.5" /></>);
export const IconChevronLeft = make(<path d="M15 6l-6 6 6 6" />);
export const IconChevronRight = make(<path d="M9 6l6 6-6 6" />);
export const IconEye = make(<><path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z" /><circle cx="12" cy="12" r="3" /></>);
export const IconEyeOff = make(<><path d="M3 3l18 18" /><path d="M10.6 5.1A10 10 0 0 1 12 5c6.4 0 10 7 10 7a17 17 0 0 1-3.2 4.1" /><path d="M6.6 6.6A16.6 16.6 0 0 0 2 12s3.6 7 10 7a9.7 9.7 0 0 0 4.4-1" /><path d="M9.9 9.9a3 3 0 0 0 4.2 4.2" /></>);
export const IconEdit = make(<><path d="M4 20h4L19 9a2.8 2.8 0 0 0-4-4L4 16v4z" /><path d="M13.5 6.5l4 4" /></>);
export const IconSwap = make(<><path d="M7 7h13l-3-3" /><path d="M17 17H4l3 3" /></>);
export const IconTrash = make(<><path d="M4 7h16" /><path d="M9 7V4h6v3" /><path d="M6 7l1 13h10l1-13" /><path d="M10 11v6" /><path d="M14 11v6" /></>);
export const IconAlert = make(<><circle cx="12" cy="12" r="9" /><path d="M12 7.5v5" /><path d="M12 16.2v.1" /></>);
export const IconCheck = make(<><circle cx="12" cy="12" r="9" /><path d="M8.2 12.3l2.6 2.6 5-5.3" /></>);
export const IconUser = make(<><circle cx="12" cy="8" r="3.8" /><path d="M4.5 20c.8-3.8 3.7-6 7.5-6s6.7 2.2 7.5 6" /></>);
export const IconBan = make(<><circle cx="12" cy="12" r="9" /><path d="M5.6 5.6l12.8 12.8" /></>);
export const IconTicket = make(<><path d="M3 9a2 2 0 0 0 0 6v3h18v-3a2 2 0 0 0 0-6V6H3v3z" /><path d="M13 6v12" strokeDasharray="2 3" /></>);
export const IconInbox = make(<><path d="M3 13l2.5-8h13L21 13" /><path d="M3 13v6h18v-6h-5l-1.5 2.5h-5L8 13H3z" /></>);
export const IconArrowRight = make(<><path d="M5 12h14" /><path d="M13 6l6 6-6 6" /></>);
export const IconSun = make(<><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></>);
export const IconMoon = make(<path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" />);
export const IconMapPin = make(<><path d="M12 21s7-6 7-11a7 7 0 0 0-14 0c0 5 7 11 7 11z" /><circle cx="12" cy="10" r="2.5" /></>);

export const IconBell = make(<><path d="M6 17V11a6 6 0 0 1 12 0v6l1.5 2h-15L6 17z" /><path d="M10 21a2 2 0 0 0 4 0" /></>);
export const IconChevronDown = make(<path d="M6 9l6 6 6-6" />);
export const IconCheckPlain = make(<path d="M5 12.5l4.5 4.5L19 7.5" />);
export const IconInfo = make(<><circle cx="12" cy="12" r="9" /><path d="M12 11v5.5" /><path d="M12 7.6v.1" /></>);

export const Logo: React.FC<{ size?: number }> = ({ size = 32 }) => (
  <span
    className="logo-mark inline-grid place-items-center shrink-0"
    style={{ width: size, height: size, borderRadius: size * 0.28, background: 'linear-gradient(140deg, var(--brand), color-mix(in srgb, var(--brand) 60%, #000))' }}
  >
    <svg viewBox="0 0 24 24" width={size * 0.58} height={size * 0.58} fill="none" stroke="var(--on-brand)" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M12 21s7-6 7-11a7 7 0 0 0-14 0c0 5 7 11 7 11z" />
      <circle cx="12" cy="10" r="2.4" fill="var(--on-brand)" stroke="none" />
    </svg>
  </span>
);

export const IconMail = make(<><rect x="3" y="5" width="18" height="14" rx="3" /><path d="M3.5 7.5l8.5 6 8.5-6" /></>);
export const IconLock = make(<><rect x="4.5" y="10.5" width="15" height="10" rx="3" /><path d="M8 10.5V8a4 4 0 0 1 8 0v2.5" /><path d="M12 15v2" /></>);
