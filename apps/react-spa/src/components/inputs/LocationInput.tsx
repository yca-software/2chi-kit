import { useState, useEffect, useCallback, useRef } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@yca-software/design-system";
import { MapPin, Loader2 } from "lucide-react";
import { useLocationAutocompleteQuery } from "@/api";
import type { Point } from "@/types/geo";

export interface LocationData {
  address: string;
  city: string;
  zip: string;
  country: string;
  placeId: string;
  geo: Point;
  timezone: string;
}

interface LocationInputProps {
  value?: string; // placeId
  displayValue?: string; // formatted address to display
  onChange?: (placeId: string | null) => void;
  onError?: (error: string) => void;
  placeholder?: string;
  disabled?: boolean;
}

export const LocationInput = ({
  value,
  displayValue,
  onChange,
  onError,
  placeholder,
  disabled = false,
}: LocationInputProps) => {
  const { t } = useTranslation("common");
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const isUserTypingRef = useRef(false);

  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");
  const [inputValue, setInputValue] = useState(displayValue ?? "");
  const [showPredictions, setShowPredictions] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);

  // Debounce search (300ms)
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearchQuery(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const autocompleteQuery = useLocationAutocompleteQuery(
    debouncedSearchQuery,
    !!debouncedSearchQuery.trim(),
  );
  const predictions = autocompleteQuery.data?.predictions ?? [];
  const isLoading = autocompleteQuery.isFetching;
  const selectableCount = predictions.length;

  // Sync displayValue / value from props
  useEffect(() => {
    if (displayValue !== undefined) setInputValue(displayValue);
  }, [displayValue]);
  useEffect(() => {
    if (!value) setInputValue("");
  }, [value]);

  useEffect(() => {
    setHighlightedIndex(-1);
  }, [predictions]);

  useEffect(() => {
    if (highlightedIndex >= 0 && listRef.current) {
      const item = listRef.current.querySelector(
        `[data-index="${highlightedIndex}"]`,
      );
      item?.scrollIntoView({ block: "nearest" });
    }
  }, [highlightedIndex]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      if (containerRef.current && !containerRef.current.contains(target)) {
        setShowPredictions(false);
        setHighlightedIndex(-1);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Report autocomplete errors
  useEffect(() => {
    if (!autocompleteQuery.error) return;
    let message = t("locationInput.noMatches");
    const err = autocompleteQuery.error;
    if (err instanceof Error) message = err.message;
    else if (err && typeof err === "object" && "message" in err)
      message = String((err as { message?: string }).message);
    onError?.(message);
  }, [autocompleteQuery.error, onError, t]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    isUserTypingRef.current = true;
    setInputValue(newValue);
    setSearchQuery(newValue);
    if (!newValue) {
      onChange?.(null);
      setShowPredictions(false);
      setTimeout(() => {
        isUserTypingRef.current = false;
      }, 100);
      return;
    }
    setShowPredictions(true);
    isUserTypingRef.current = false;
  };

  const handleFocus = () => {
    if (inputValue.trim().length > 0) {
      setShowPredictions(true);
      setSearchQuery(inputValue.trim());
    }
  };

  const selectPlace = useCallback(
    (placeId: string, description: string) => {
      setInputValue(description);
      setShowPredictions(false);
      setHighlightedIndex(-1);
      onChange?.(placeId);
    },
    [onChange],
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!showPredictions || selectableCount === 0) {
        if (e.key === "Escape") {
          e.preventDefault();
          setShowPredictions(false);
          setHighlightedIndex(-1);
        }
        return;
      }
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev < selectableCount - 1 ? prev + 1 : 0,
        );
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev <= 0 ? selectableCount - 1 : prev - 1,
        );
        return;
      }
      if (
        e.key === "Enter" &&
        highlightedIndex >= 0 &&
        predictions[highlightedIndex]
      ) {
        e.preventDefault();
        const pred = predictions[highlightedIndex];
        selectPlace(pred.placeId, pred.description);
        return;
      }
      if (e.key === "Escape") {
        e.preventDefault();
        setShowPredictions(false);
        setHighlightedIndex(-1);
      }
    },
    [
      showPredictions,
      selectableCount,
      highlightedIndex,
      predictions,
      selectPlace,
    ],
  );

  return (
    <div ref={containerRef} className="relative">
      <div className="relative">
        {isLoading && (
          <div className="absolute right-3 top-1/2 z-10 -translate-y-1/2">
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          </div>
        )}
        <Input
          ref={inputRef}
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder ?? t("locationInput.placeholder")}
          disabled={disabled}
          className="pl-10"
          onFocus={handleFocus}
          aria-autocomplete="list"
          aria-expanded={showPredictions && selectableCount > 0}
          aria-controls={
            showPredictions ? "location-predictions-list" : undefined
          }
          aria-activedescendant={
            highlightedIndex >= 0
              ? `location-prediction-${highlightedIndex}`
              : undefined
          }
          role="combobox"
        />
        <MapPin className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      </div>
      {showPredictions && inputValue.trim().length > 0 && (
        <div
          ref={listRef}
          id="location-predictions-list"
          role="listbox"
          className="location-predictions absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md border bg-background shadow-lg"
        >
          {isLoading && predictions.length === 0 ? (
            <div className="flex items-center gap-2 px-4 py-3 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 shrink-0 animate-spin" />
              {t("locationInput.searching")}
            </div>
          ) : (
            predictions.map((prediction, index) => (
              <button
                key={prediction.placeId}
                type="button"
                role="option"
                id={`location-prediction-${index}`}
                data-index={index}
                aria-selected={highlightedIndex === index}
                className={`w-full cursor-pointer px-4 py-2 text-left transition-colors ${
                  highlightedIndex === index
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent hover:text-accent-foreground"
                }`}
                onMouseEnter={() => setHighlightedIndex(index)}
                onClick={() =>
                  selectPlace(prediction.placeId, prediction.description)
                }
              >
                <div className="font-medium">
                  {prediction.structuredFormatting?.mainText ??
                    prediction.description}
                </div>
                {prediction.structuredFormatting?.secondaryText && (
                  <div className="text-sm text-muted-foreground">
                    {prediction.structuredFormatting.secondaryText}
                  </div>
                )}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
};
