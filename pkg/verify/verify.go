// Updated cmd to include a timeout of 5 minutes
cmd := exec.Command("go", "test", "./...", "-cover", "-timeout", "5m")
