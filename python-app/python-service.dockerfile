# Use the official Python image from the Docker registry
FROM python:3.12

# Set the working directory inside the container
WORKDIR /app/

RUN apt-get update && apt-get install -y \
    libgl1-mesa-glx \
    libglib2.0-0

# Copy requirements.txt first (for better caching)
COPY requirements.txt /app/

# Install dependencies globally
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application code into the container
COPY . /app/

# Expose the port that the Flask app will run on
EXPOSE 8000

# Ensure the virtual environment is activated before running the app
CMD ["python", "main.py"]
