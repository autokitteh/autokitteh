import hashlib
import os
import log


def get_digest(data):
    s = hashlib.sha256()
    s.update(data)
    return s.hexdigest()


class LargeObjectsManger:
    def __init__(self, path, enabled=False):
        self.cache = {}
        self.path = path
        self.enabled = enabled

    def set(self, val):
        digest = get_digest(val)
        self.cache[digest] = val
        self.persist(digest, val)
        return digest

    def get(self, key):
        log.info(f"try fetch {key}")
        if key not in self.cache:
            log.info("not found in cache")
            data = self.load(key)
            log.info("loaded")
            self.cache[key] = data
        log.info("returning from cache")
        return self.cache.get(key)  # Default None

    def local_path(self, key):
        return os.path.join(self.path, key)

    def persist(self, key, val):
        file_name = self.local_path(key)
        log.info(f"writing to {file_name}")
        with open(file_name, "wb") as file_object:
            file_object.write(val)

    def load(self, key):
        file_name = self.local_path(key)
        log.info(f"reading from {file_name}")
        with open(file_name, "rb") as file_object:
            return file_object.read()
