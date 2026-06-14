export type TabItem = {
  key: string;
  label: string;
};

export function Tabs({
  tabs,
  active,
  onChange,
}: {
  tabs: TabItem[];
  active: string;
  onChange: (key: string) => void;
}) {
  return (
    <div className="border-b flex flex-wrap gap-1">
      {tabs.map((tab) => (
        <button
          key={tab.key}
          type="button"
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px ${
            active === tab.key
              ? "border-black text-black"
              : "border-transparent text-gray-500 hover:text-gray-700"
          }`}
          onClick={() => onChange(tab.key)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
