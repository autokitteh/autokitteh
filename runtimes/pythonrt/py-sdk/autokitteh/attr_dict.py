"""Helper class to allow attribute access to dictionary keys."""


class AttrDict(dict):
    """Allow attribute access to dictionary keys.

    >>> config = AttrDict({'server': {'port': 8080}, 'debug': True})
    >>> config.server.port
    8080
    >>> config.debug
    True
    """

    def __getattr__(self, name: str):
        try:
            value = self[name]
            if isinstance(value, dict):
                value = AttrDict(value)
            return value
        except KeyError:
            # "from None" will remove the confusing KeyError from the stack trace
            raise AttributeError(name) from None

    def __setattr__(self, name: str, value):
        # The default __getattr__ doesn't fail but also don't change values.
        cls = self.__class__.__name__
        raise NotImplementedError(f"{cls} does not support setting attributes")

    def __dir__(self):
        return self.keys()
