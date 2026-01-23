import sys
import json
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer, pipeline

# Set up device
device = "cuda" if torch.cuda.is_available() else "cpu"

def main():
    # Basic conversation pipeline
    model_id = "microsoft/phi-2" # Lightweight default
    
    sys.stderr.write(f"[Python] Loading model {model_id} on {device}...\n")
    sys.stderr.flush()

    try:
        # Load model & tokenizer
        tokenizer = AutoTokenizer.from_pretrained(model_id, trust_remote_code=True)
        model = AutoModelForCausalLM.from_pretrained(model_id, torch_dtype=torch.float16 if device=="cuda" else torch.float32, trust_remote_code=True).to(device)
        
        sys.stderr.write("[Python] Model loaded ready.\n")
        sys.stderr.flush()

        # Listen for lines
        for line in sys.stdin:
            if not line: break
            
            try:
                data = json.loads(line)
                prompt = data.get("prompt", "")
                
                # Simple generation
                inputs = tokenizer(prompt, return_tensors="pt", return_attention_mask=False).to(device)
                
                outputs = model.generate(**inputs, max_length=200)
                text = tokenizer.batch_decode(outputs)[0]
                
                # Clean up prompt from output if needed, though phi-2 usually continues
                # Sending back JSON
                print(json.dumps({"response": text}))
                sys.stdout.flush()
                
            except Exception as e:
                sys.stderr.write(f"[Python Error] {e}\n")
                print(json.dumps({"error": str(e)}))
                sys.stdout.flush()

    except Exception as e:
        sys.stderr.write(f"[Python Critical Error] {e}\n")
        sys.exit(1)

if __name__ == "__main__":
    main()
