import { useContext, useState } from "react";
import styled from "@emotion/styled";
import { FaSave, FaChevronDown, FaChevronUp } from "react-icons/fa";
import { Item } from "../../../types";
import { LAND_COLOR } from "../../../utils/colors.ts";
import { saveNote } from "../../../api/notesApi.ts";
import { AuthContext } from "../../../contexts/Auth.ts";

const NotesSection = styled.div`
  margin-top: 1rem;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  background-color: #f9f9f9;
`;

const NotesHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  cursor: pointer;
  user-select: none;
  border-bottom: ${({ expanded }: { expanded: boolean }) =>
    expanded ? "1px solid #e0e0e0" : "none"};

  &:hover {
    background-color: #f0f0f0;
  }
`;

const NotesTitle = styled.div`
  font-size: 0.9rem;
  font-weight: bold;
  color: #333;
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

const NotesContent = styled.div`
  padding: 1rem;
  display: ${({ expanded }: { expanded: boolean }) =>
    expanded ? "block" : "none"};
`;

const NotesTextarea = styled.textarea`
  width: 100%;
  background: white;
  color: black;
  min-height: 120px;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  resize: vertical;
  font-family: inherit;
  font-size: 0.9rem;
  box-sizing: border-box;
`;

const SaveButton = styled.button`
  margin-top: 0.5rem;
  padding: 0.4rem 0.8rem;
  background-color: ${LAND_COLOR};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;

  &:hover {
    opacity: 0.8;
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const NotePreview = styled.span`
  font-size: 0.7rem;
  color: #666;
  font-weight: normal;
`;

interface NotesEditorProps {
  item: Item;
}

export const NotesEditor = ({ item }: NotesEditorProps) => {
  const [notes, setNotes] = useState(item.notes || "");
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const { token } = useContext(AuthContext);

  const handleSave = async () => {
    if (!token) {
      return;
    }
    setSaving(true);
    try {
      await saveNote(item.key, {
        note: notes,
      });

      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (error) {
      console.error("Failed to save notes:", error);
    } finally {
      setSaving(false);
    }
  };

  if (!token) {
    return null;
  }

  return (
    <NotesSection>
      <NotesHeader expanded={expanded} onClick={() => setExpanded(!expanded)}>
        <NotesTitle>
          <span>Notes</span>
          {notes && !expanded && (
            <NotePreview>
              ({notes.length > 50 ? `${notes.substring(0, 50)}...` : notes})
            </NotePreview>
          )}
        </NotesTitle>
        {expanded ? <FaChevronUp /> : <FaChevronDown />}
      </NotesHeader>
      <NotesContent expanded={expanded}>
        <NotesTextarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Add notes for this item..."
        />
        <SaveButton
          onClick={handleSave}
          disabled={saving || notes === (item.notes || "")}
        >
          <FaSave />
          {saving ? "Saving..." : saved ? "Saved!" : "Save Notes"}
        </SaveButton>
      </NotesContent>
    </NotesSection>
  );
};
