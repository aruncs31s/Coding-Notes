```cpp
#include <iostream>
using namespace std;

class Base {
public:
    virtual void display() {
        cout << "Base class" << endl;
    }
};

class Derived : public Base {
public:
    void display() override {
        cout << "Derived class" << endl;
    }
};

int main() {
    Base* ptr = new Derived();
    ptr->display();   // Calls Derived::display()
}
```
Op:
```
Derived class
```